package index

import (
	"strings"

	"github.com/GabrielMoody/Cariin/internal/documents"
)

func tokenize(text string) []string {
	text = strings.ToLower(text)

	replacer := strings.NewReplacer(
		".", " ",
		",", " ",
		"!", " ",
		"?", " ",
		":", " ",
		";", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"/", " ",
		"-", " ",
		"_", " ",
	)

	text = replacer.Replace(text)

	return strings.Fields(text)
}

func Build(docs []documents.Document) *InvertedIdx {
	idx := make(InvertedIdx)

	for _, v := range docs {
		text := v.Title + " " + v.Body

		tokens := tokenize(text)
		seen := make(map[string]bool)

		for _, token := range tokens {
			if seen[token] {
				continue
			}

			idx[token] = append(idx[token], v.Id)
			seen[token] = true
		}
	}

	return &idx
}
