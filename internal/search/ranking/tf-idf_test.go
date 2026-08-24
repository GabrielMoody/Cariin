package ranking

import (
	"reflect"
	"testing"

	"github.com/GabrielMoody/Cariin/internal/index"
	ast "github.com/GabrielMoody/Cariin/internal/search/AST"
)

func TestRankTFIDFOrdersByTermFrequency(t *testing.T) {
	all := index.InvertedIdx{
		"go": {
			{DocId: 1, Positions: []int{0, 1, 2}},
			{DocId: 2, Positions: []int{0}},
		},
	}
	query := ast.TermQuery{Term: "go"}

	got := RankTFIDF(query, index.Index{All: &all}, []index.Posting{all["go"][1], all["go"][0]})
	ids := []int64{got[0].DocId, got[1].DocId}
	if !reflect.DeepEqual(ids, []int64{1, 2}) {
		t.Fatalf("RankTFIDF() IDs = %v, want [1 2]", ids)
	}
}

func TestRankTFIDFUsesFieldIndex(t *testing.T) {
	all := index.InvertedIdx{"go": {{DocId: 1}, {DocId: 2}}}
	title := index.InvertedIdx{"go": {{DocId: 2, Positions: []int{0, 1}}}}
	query := ast.TermQuery{Field: "title", Term: "go"}

	got := RankTFIDF(query, index.Index{All: &all, Title: &title}, []index.Posting{{DocId: 2}, {DocId: 1}})
	if got[0].DocId != 2 {
		t.Fatalf("RankTFIDF() first ID = %d, want 2", got[0].DocId)
	}
}
