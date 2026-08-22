package parser

import (
	"reflect"
	"testing"

	"github.com/GabrielMoody/Cariin/internal/index"
	ast "github.com/GabrielMoody/Cariin/internal/search/AST"
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

	gotPostings := query.Evaluate(index.Index{
		All: &invertedIndex,
	})
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

	gotPostings := query.Evaluate(index.Index{
		All: &index.InvertedIdx{
			"go":   {{DocId: 1}},
			"rust": {{DocId: 2}},
			"web":  {{DocId: 3}},
		},
	})
	got := postingIDs(gotPostings)
	want := []int64{1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %v, want %v", got, want)
	}
}

func TestParseFieldQuery(t *testing.T) {
	query, err := Parse("title:go")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	term, ok := query.(ast.TermQuery)
	if !ok {
		t.Fatalf("Parse() returned %T, want ast.TermQuery", query)
	}
	if term.Field != "title" || term.Term != "go" {
		t.Fatalf("Parse() = %#v, want title field and go term", term)
	}
}

func TestEvaluateFieldQuery(t *testing.T) {
	query, err := Parse("title:go")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	titleIndex := index.InvertedIdx{"go": {{DocId: 1}}}
	bodyIndex := index.InvertedIdx{"go": {{DocId: 2}}}
	allIndex := index.InvertedIdx{"go": {{DocId: 1}, {DocId: 2}}}
	got := postingIDs(query.Evaluate(index.Index{
		All:   &allIndex,
		Title: &titleIndex,
		Body:  &bodyIndex,
	}))
	if !reflect.DeepEqual(got, []int64{1}) {
		t.Fatalf("Evaluate() = %v, want [1]", got)
	}
}

func TestParseRejectsInvalidFieldQuery(t *testing.T) {
	for _, input := range []string{"author:go", "title:", ":go"} {
		if _, err := Parse(input); err == nil {
			t.Errorf("Parse(%q) returned nil error", input)
		}
	}
}

func TestParseAndEvaluatePhrase(t *testing.T) {
	query, err := Parse("go programming")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	got := postingIDs(query.Evaluate(index.Index{
		All: &index.InvertedIdx{
			"go":          {{DocId: 1, Positions: []int{0}}, {DocId: 2, Positions: []int{0}}},
			"programming": {{DocId: 1, Positions: []int{1}}, {DocId: 2, Positions: []int{3}}},
		}}))
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

	got := postingIDs(query.Evaluate(index.Index{
		All: &index.InvertedIdx{
			"go":    {{DocId: 1, Positions: []int{0}}, {DocId: 2, Positions: []int{0}}},
			"redis": {{DocId: 1, Positions: []int{1}}, {DocId: 2, Positions: []int{4}}},
			"sql":   {{DocId: 1, Positions: []int{9}}, {DocId: 2, Positions: []int{5}}},
		}}))
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
