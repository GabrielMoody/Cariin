package token

import "testing"

func TestLexGroupsWordsIntoPhraseTerms(t *testing.T) {
	tokens, err := Lex("go programming AND (rust language OR python)")
	if err != nil {
		t.Fatalf("Lex() error = %v", err)
	}

	want := []Token{
		{Type: TokenTerm, Value: "go programming"},
		{Type: TokenAND, Value: "AND"},
		{Type: TokenLeftParen, Value: "("},
		{Type: TokenTerm, Value: "rust language"},
		{Type: TokenOR, Value: "OR"},
		{Type: TokenTerm, Value: "python"},
		{Type: TokenRightParen, Value: ")"},
		{Type: TokenEOF},
	}

	if len(tokens) != len(want) {
		t.Fatalf("Lex() returned %d tokens, want %d: %v", len(tokens), len(want), tokens)
	}
	for i := range want {
		if tokens[i] != want[i] {
			t.Errorf("token %d = %#v, want %#v", i, tokens[i], want[i])
		}
	}
}
