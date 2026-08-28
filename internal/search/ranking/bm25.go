package ranking

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/GabrielMoody/Cariin/internal/documents"
	"github.com/GabrielMoody/Cariin/internal/index"
	ast "github.com/GabrielMoody/Cariin/internal/search/AST"
)

type BM25 struct {
	TotalDocuments float64
	Avgdl          float64
	K1             float64
	B              float64
}

func NewBM25() *BM25 {
	totalLength := 0

	for _, doc := range documents.Documents {
		totalLength += len(strings.Fields(doc.Body))
		totalLength += len(strings.Fields(doc.Title))
	}

	return &BM25{
		TotalDocuments: float64(len(documents.Documents)),
		Avgdl:          float64(totalLength) / float64(len(documents.Documents)),
		K1:             1.2,
		B:              0.75,
	}
}

func (b *BM25) RankBM25(query ast.Query, idx index.Index, postings []index.Posting) []index.Posting {
	scores := make(map[int64]float64, len(postings))
	for _, posting := range postings {
		scores[posting.DocId] = 0
	}

	b.addBMTermScores(query, idx, scores)

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
	fmt.Println(scores)
	return result
}

func (b *BM25) addBMTermScores(query ast.Query, idx index.Index, scores map[int64]float64) {
	switch query := query.(type) {
	case ast.TermQuery:
		fieldIndex := selectRankingFieldIndex(idx, query.Field)
		for _, term := range strings.Fields(strings.ToLower(query.Term)) {
			postings := fieldIndex[term]
			if len(postings) == 0 {
				continue
			}

			idf := b.inverseDocumentFrequency(fieldIndex, term)
			for _, posting := range postings {
				if _, candidate := scores[posting.DocId]; candidate {
					fq := float64(len(posting.Positions))
					doc := documents.Documents[posting.DocId]
					docLength :=
						float64(len(strings.Fields(doc.Body))) +
							float64(len(strings.Fields(doc.Title)))

					tf := (fq * (b.K1 + 1) / (fq + b.K1*(1-b.B+b.B*(docLength/b.Avgdl))))
					scores[posting.DocId] += tf * idf
				}
			}
		}
	case ast.AndQuery:
		b.addBMTermScores(query.Left, idx, scores)
		b.addBMTermScores(query.Right, idx, scores)
	case ast.OrQuery:
		b.addBMTermScores(query.Left, idx, scores)
		b.addBMTermScores(query.Right, idx, scores)
	case ast.NotQuery:
		return
	}
}

func (b *BM25) inverseDocumentFrequency(fieldIndex index.InvertedIdx, term string) float64 {

	df := len(fieldIndex[term])

	if df == 0 {
		return 0
	}

	return math.Log(1 + (b.TotalDocuments-float64(df)+0.5)/(float64(df)+0.5))
}
