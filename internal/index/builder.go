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

func Build(docs map[int64]documents.Document, field func(documents.Document) string) *InvertedIdx {
	idx := make(InvertedIdx)

	for _, v := range docs {
		tokens := tokenize(field(v))

		for i, token := range tokens {
			postings := idx[token]
			postingFound := false

			for postingIndex := range postings {
				if postings[postingIndex].DocId != v.Id {
					continue
				}

				postings[postingIndex].Positions = append(postings[postingIndex].Positions, i)
				postingFound = true
				break
			}

			if !postingFound {
				postings = append(postings, Posting{
					DocId:     v.Id,
					Positions: []int{i},
				})
			}

			idx[token] = postings
		}
	}

	return &idx
}
