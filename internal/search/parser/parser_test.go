package parser

import (
	"reflect"
	"testing"

	"github.com/GabrielMoody/Cariin/internal/index"
)

func TestParseAndEvaluateBooleanQuery(t *testing.T) {
	invertedIndex := index.InvertedIdx{
		"go":     {1, 2},
		"rust":   {3},
		"python": {2},
		"web":    {1, 3},
	}

	query, err := Parse("(GO OR rust) AND NOT python")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	got := query.Evaluate(invertedIndex)
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

	got := query.Evaluate(index.InvertedIdx{
		"go":   {1},
		"rust": {2},
		"web":  {3},
	})
	want := []int64{1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %v, want %v", got, want)
	}
}

func TestParseRejectsMalformedQueries(t *testing.T) {
	for _, input := range []string{"", "go AND", "(go OR rust", "go OR OR rust"} {
		if _, err := Parse(input); err == nil {
			t.Errorf("Parse(%q) returned nil error", input)
		}
	}
}
