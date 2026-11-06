package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/temoto/robotstxt"
)

func (c *Crawler) allowedByRobots(
	ctx context.Context,
	target *url.URL,
) (bool, error) {

	host := target.Host

	state := c.getDomainState(host)

	state.mu.Lock()
	defer state.mu.Unlock()

	// Cache robots.txt for 1 hour.
	if state.robots != nil &&
		time.Since(state.robotsTime) < time.Hour {

		group := state.robots.FindGroup(c.userAgent)

		if group == nil {
			return true, nil
		}

		return group.Test(target.Path), nil
	}

	robotsURL := fmt.Sprintf(
		"%s://%s/robots.txt",
		target.Scheme,
		target.Host,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		robotsURL,
		nil,
	)

	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	// A missing robots.txt does not mean crawling is forbidden.
	if resp.StatusCode == http.StatusNotFound {
		state.robots = nil
		state.robotsTime = time.Now()

		return true, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Conservative behavior for robots.txt failures.
		return false, fmt.Errorf(
			"robots.txt returned status %d",
			resp.StatusCode,
		)
	}

	data, err := io.ReadAll(
		io.LimitReader(resp.Body, 512*1024),
	)

	if err != nil {
		return false, err
	}

	robots, err := robotstxt.FromBytes(data)

	if err != nil {
		return false, err
	}

	state.robots = robots
	state.robotsTime = time.Now()

	group := robots.FindGroup(c.userAgent)

	if group == nil {
		return true, nil
	}

	return group.Test(target.Path), nil
}
