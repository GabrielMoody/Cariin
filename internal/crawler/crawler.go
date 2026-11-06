package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
	"golang.org/x/time/rate"
)

const (
	defaultUserAgent = "CariinBot/1.0 (Learning Project)"

	maxBodySize     = 5 * 1024 * 1024
	maxRetries      = 5
	requestTimeout  = 20 * time.Second
	requestInterval = 2 * time.Second
)

type Document struct {
	Title     string
	Domain    string
	URL       string
	Content   string
	CrawledAt time.Time
}

type domainState struct {
	limiter *rate.Limiter

	robots     *robotstxt.RobotsData
	robotsTime time.Time

	mu sync.Mutex
}

type Crawler struct {
	client    *http.Client
	userAgent string

	mu      sync.Mutex
	domains map[string]*domainState

	visited map[string]struct{}
}

func New() *Crawler {
	return &Crawler{
		client: &http.Client{
			Timeout: requestTimeout,
			CheckRedirect: func(
				req *http.Request,
				via []*http.Request,
			) error {
				if len(via) >= 10 {
					return errors.New("too many redirects")
				}

				return nil
			},
		},

		userAgent: defaultUserAgent,

		domains: make(map[string]*domainState),
		visited: make(map[string]struct{}),
	}
}

func (c *Crawler) Crawl(ctx context.Context, rawURL string) (*Document, error) {

	target, err := url.Parse(rawURL)

	if err != nil {
		return nil, err
	}

	if target.Scheme != "http" &&
		target.Scheme != "https" {

		return nil, fmt.Errorf(
			"unsupported scheme: %s",
			target.Scheme,
		)
	}

	// Normalize.
	target.Fragment = ""

	// Check robots.txt BEFORE fetching.
	allowed, err := c.allowedByRobots(
		ctx,
		target,
	)

	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, fmt.Errorf(
			"blocked by robots.txt: %s",
			target.String(),
		)
	}

	// Deduplicate.
	if !c.markVisited(target.String()) {
		return nil, fmt.Errorf(
			"already visited: %s",
			target.String(),
		)
	}

	var resp *http.Response

	for attempt := 0; attempt < maxRetries; attempt++ {

		resp, err = c.fetch(ctx, target)

		if err != nil {

			if attempt == maxRetries-1 {
				return nil, err
			}

			wait := backoff(attempt)

			log.Printf(
				"request failed: %s; retrying in %s",
				err,
				wait,
			)

			if err := sleepContext(ctx, wait); err != nil {
				return nil, err
			}

			continue
		}

		switch {

		case resp.StatusCode == http.StatusTooManyRequests:

			wait := retryAfter(resp)

			resp.Body.Close()

			if wait == 0 {
				wait = backoff(attempt)
			}

			log.Printf(
				"rate limited by %s; waiting %s",
				target.Host,
				wait,
			)

			if err := sleepContext(ctx, wait); err != nil {
				return nil, err
			}

			continue

		case resp.StatusCode >= 500:

			resp.Body.Close()

			wait := backoff(attempt)

			log.Printf(
				"server error %d; retrying in %s",
				resp.StatusCode,
				wait,
			)

			if err := sleepContext(ctx, wait); err != nil {
				return nil, err
			}

			continue

		case resp.StatusCode != http.StatusOK:

			resp.Body.Close()

			return nil, fmt.Errorf(
				"HTTP status: %d",
				resp.StatusCode,
			)
		}

		break
	}

	if resp == nil {
		return nil, errors.New("no response")
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	if !strings.Contains(
		strings.ToLower(contentType),
		"text/html",
	) {
		return nil, fmt.Errorf(
			"unsupported content type: %s",
			contentType,
		)
	}

	body := io.LimitReader(
		resp.Body,
		maxBodySize,
	)

	html, err := io.ReadAll(body)

	if err != nil {
		return nil, err
	}

	return parseDocument(
		target,
		html,
	)
}

func (c *Crawler) getDomainState(host string) *domainState {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, exists := c.domains[host]

	if exists {
		return state
	}

	state = &domainState{
		limiter: rate.NewLimiter(
			rate.Every(requestInterval),
			1,
		),
	}

	c.domains[host] = state

	return state
}
