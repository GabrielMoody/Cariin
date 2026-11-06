package crawler

import (
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
)

func parseDocument(target *url.URL, html []byte) (*Document, error) {

	article, err := readability.FromReader(
		strings.NewReader(string(html)),
		target,
	)

	if err != nil {
		return nil, err
	}

	return &Document{
		Title:     strings.TrimSpace(article.Title),
		Domain:    target.Hostname(),
		URL:       target.String(),
		Content:   strings.TrimSpace(article.TextContent),
		CrawledAt: time.Now().UTC(),
	}, nil
}
