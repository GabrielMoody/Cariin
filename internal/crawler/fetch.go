package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func Fetch(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)

	req.Header.Set(
		"User-Agent",
		"CariinSearchEngine/1.0 (learning project)",
	)

	req.Header.Set(
		"Accept",
		"text/html,application/xhtml+xml",
	)

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"unexpected status: %d",
			resp.StatusCode,
		)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("error %v\n", err)
		return "", err
	}
	return string(body), nil
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
