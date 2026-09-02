package crawler

import (
	"net/url"
	"strings"
)

func NormalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)

	if err != nil {
		return "", err
	}

	u.Fragment = ""

	u.Host = strings.ToLower(u.Host)

	return u.String(), nil
}

func ResolveURL(base string, href string) (string, error) {

	baseURL, err := url.Parse(base)

	if err != nil {
		return "", err
	}

	refURL, err := url.Parse(href)

	if err != nil {
		return "", err
	}

	return baseURL.ResolveReference(refURL).String(), nil
}
