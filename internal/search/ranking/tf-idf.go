package ranking

import (
	"math"
	"sort"
	"strings"

	"github.com/GabrielMoody/Cariin/internal/index"
	ast "github.com/GabrielMoody/Cariin/internal/search/AST"
)

type scoredPosting struct {
	posting index.Posting
	score   float64
}

func TFIDF(query ast.Query, idx index.Index, postings []index.Posting) []index.Posting {
	scores := make(map[int64]float64, len(postings))
	for _, posting := range postings {
		scores[posting.DocId] = 0
	}

	addTermScores(query, idx, scores)

	ranked := make([]scoredPosting, 0, len(postings))
	for _, posting := range postings {
		ranked = append(ranked, scoredPosting{
			posting: posting,
			score:   scores[posting.DocId],
		})
	}

	sort.SliceStable(ranked, func(left, right int) bool {
		if ranked[left].score == ranked[right].score {
			return ranked[left].posting.DocId < ranked[right].posting.DocId
		}
		return ranked[left].score > ranked[right].score
	})

	result := make([]index.Posting, 0, len(ranked))
	for _, scored := range ranked {
		result = append(result, scored.posting)
	}
	return result
}

func addTermScores(query ast.Query, idx index.Index, scores map[int64]float64) {
	switch query := query.(type) {
	case ast.TermQuery:
		fieldIndex := selectRankingFieldIndex(idx, query.Field)
		for _, term := range strings.Fields(strings.ToLower(query.Term)) {
			postings := fieldIndex[term]
			if len(postings) == 0 {
				continue
			}

			idf := inverseDocumentFrequency(fieldIndex, term)
			for _, posting := range postings {
				if _, candidate := scores[posting.DocId]; candidate {
					tf := float64(len(posting.Positions))
					scores[posting.DocId] += tf * idf
				}
			}
		}
	case ast.AndQuery:
		addTermScores(query.Left, idx, scores)
		addTermScores(query.Right, idx, scores)
	case ast.OrQuery:
		addTermScores(query.Left, idx, scores)
		addTermScores(query.Right, idx, scores)
	case ast.NotQuery:
		return
	}
}

func inverseDocumentFrequency(fieldIndex index.InvertedIdx, term string) float64 {
	documents := make(map[int64]struct{})
	for _, postings := range fieldIndex {
		for _, posting := range postings {
			documents[posting.DocId] = struct{}{}
		}
	}
	documentCount := len(documents)
	if documentCount == 0 {
		return 0
	}

	documentFrequency := len(fieldIndex[term])
	return math.Log(float64(documentCount) / float64(documentFrequency))
}

func selectRankingFieldIndex(idx index.Index, field string) index.InvertedIdx {
	var selected *index.InvertedIdx
	switch field {
	case "title":
		selected = idx.Title
	case "body":
		selected = idx.Body
	case "url":
		selected = idx.Url
	default:
		selected = idx.All
	}

	if selected == nil {
		return index.InvertedIdx{}
	}
	return *selected
}
