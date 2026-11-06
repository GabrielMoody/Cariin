package crawler

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (c *Crawler) fetch(ctx context.Context, target *url.URL) (*http.Response, error) {

	state := c.getDomainState(target.Host)

	if err := state.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		target.String(),
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set(
		"Accept",
		"text/html,application/xhtml+xml",
	)

	return c.client.Do(req)
}

func ExtractLinks(baseURL string, html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(
		strings.NewReader(html),
	)

	if err != nil {
		return nil, err
	}

	var links []string

	doc.Find("a[href]").Each(
		func(_ int, s *goquery.Selection) {

			href, exists := s.Attr("href")

			if exists {
				links = append(links, href)
			}
		},
	)

	return links, nil
}
