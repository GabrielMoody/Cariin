package crawler

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

func retryAfter(resp *http.Response) time.Duration {
	value := resp.Header.Get("Retry-After")

	if value == "" {
		return 0
	}

	// Retry-After: 120
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Retry-After can also be an HTTP date.
	if t, err := http.ParseTime(value); err == nil {
		wait := time.Until(t)

		if wait > 0 {
			return wait
		}
	}

	return 0
}

func backoff(attempt int) time.Duration {
	const (
		base = 2 * time.Second
		max  = 2 * time.Minute
	)

	wait := base * time.Duration(1<<attempt)

	if wait > max {
		return max
	}

	return wait
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)

	defer timer.Stop()

	select {
	case <-timer.C:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
