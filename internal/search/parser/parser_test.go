package parser

import (
	"reflect"
	"testing"

	"github.com/GabrielMoody/Cariin/internal/index"
)

func TestParseAndEvaluateBooleanQuery(t *testing.T) {
	invertedIndex := index.InvertedIdx{
		"go":     {{DocId: 1}, {DocId: 2}},
		"rust":   {{DocId: 3}},
		"python": {{DocId: 2}},
		"web":    {{DocId: 1}, {DocId: 3}},
	}

	query, err := Parse("(GO OR rust) AND NOT python")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	gotPostings := query.Evaluate(invertedIndex)
	got := postingIDs(gotPostings)
	want := []int64{1, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %v, want %v", got, want)
	}
}

func TestParseUsesBooleanPrecedence(t *testing.T) {
	query, err := Parse("go OR rust AND web")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	gotPostings := query.Evaluate(index.InvertedIdx{
		"go":   {{DocId: 1}},
		"rust": {{DocId: 2}},
		"web":  {{DocId: 3}},
	})
	got := postingIDs(gotPostings)
	want := []int64{1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %v, want %v", got, want)
	}
}

func TestParseAndEvaluatePhrase(t *testing.T) {
	query, err := Parse("go programming")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	got := postingIDs(query.Evaluate(index.InvertedIdx{
		"go":          {{DocId: 1, Positions: []int{0}}, {DocId: 2, Positions: []int{0}}},
		"programming": {{DocId: 1, Positions: []int{1}}, {DocId: 2, Positions: []int{3}}},
	}))
	want := []int64{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %v, want %v", got, want)
	}
}

func TestParseAndEvaluatePhraseAndImplicitAnd(t *testing.T) {
	query, err := Parse("go redis sql")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	got := postingIDs(query.Evaluate(index.InvertedIdx{
		"go":    {{DocId: 1, Positions: []int{0}}, {DocId: 2, Positions: []int{0}}},
		"redis": {{DocId: 1, Positions: []int{1}}, {DocId: 2, Positions: []int{4}}},
		"sql":   {{DocId: 1, Positions: []int{9}}, {DocId: 2, Positions: []int{5}}},
	}))
	want := []int64{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %v, want %v", got, want)
	}
}

func postingIDs(postings []index.Posting) []int64 {
	ids := make([]int64, 0, len(postings))
	for _, posting := range postings {
		ids = append(ids, posting.DocId)
	}
	return ids
}

func TestParseRejectsMalformedQueries(t *testing.T) {
	for _, input := range []string{"", "go AND", "(go OR rust", "go OR OR rust"} {
		if _, err := Parse(input); err == nil {
			t.Errorf("Parse(%q) returned nil error", input)
		}
	}
}
