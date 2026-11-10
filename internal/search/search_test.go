package search

import (
	"reflect"
	"testing"

	"github.com/GabrielMoody/Cariin/internal/documents"
	"github.com/GabrielMoody/Cariin/internal/reranker"
)

func TestDocumentsByRank(t *testing.T) {
	docs := []documents.Document{
		{Id: 1, Title: "first"},
		{Id: 2, Title: "second"},
		{Id: 3, Title: "third"},
	}
	results := []reranker.RankedDocument{
		{ID: 2, Rank: 2},
		{ID: 3, Rank: 3},
		{ID: 1, Rank: 1},
		{ID: 99, Rank: 4},
	}

	got := documentsByRank(results, docs)
	want := []documents.Document{docs[0], docs[1], docs[2]}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("documentsByRank() = %#v, want %#v", got, want)
	}
}
