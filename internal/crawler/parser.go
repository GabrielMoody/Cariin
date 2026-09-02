package crawler

import (
	"strings"

	"github.com/GabrielMoody/Cariin/internal/documents"
	"github.com/PuerkitoBio/goquery"
)

func ParseDocument(url string, html string) (*documents.T_Document, error) {
	doc, err := goquery.NewDocumentFromReader(
		strings.NewReader(html),
	)

	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(
		doc.Find("title").First().Text(),
	)

	doc.Find("script, style, nav, footer").Remove()

	body := strings.TrimSpace(
		doc.Find("body").Text(),
	)

	return &documents.T_Document{
		URL:   url,
		Title: title,
		Body:  body,
	}, nil
}
