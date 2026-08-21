package token

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenTerm TokenType = iota
	TokenAND
	TokenOR
	TokenNOT
	TokenLeftParen
	TokenRightParen
	TokenEOF
)

type Token struct {
	Type  TokenType
	Value string
}

func Lex(input string) ([]Token, error) {
	var tokens []Token
	var phrase []string
	runes := []rune(input)
	flushPhrase := func() {
		if len(phrase) == 0 {
			return
		}
		tokens = append(tokens, Token{Type: TokenTerm, Value: strings.Join(phrase, " ")})
		phrase = nil
	}

	for position := 0; position < len(runes); {
		if unicode.IsSpace(runes[position]) {
			position++
			continue
		}

		switch runes[position] {
		case '(':
			flushPhrase()
			tokens = append(tokens, Token{Type: TokenLeftParen, Value: "("})
			position++
			continue
		case ')':
			flushPhrase()
			tokens = append(tokens, Token{Type: TokenRightParen, Value: ")"})
			position++
			continue
		}

		start := position
		for position < len(runes) && !unicode.IsSpace(runes[position]) && runes[position] != '(' && runes[position] != ')' {
			position++
		}
		value := string(runes[start:position])

		switch strings.ToUpper(value) {
		case "AND":
			flushPhrase()
			tokens = append(tokens, Token{Type: TokenAND, Value: value})
		case "OR":
			flushPhrase()
			tokens = append(tokens, Token{Type: TokenOR, Value: value})
		case "NOT":
			flushPhrase()
			tokens = append(tokens, Token{Type: TokenNOT, Value: value})
		default:
			if value == "" {
				return nil, fmt.Errorf("empty term at position %d", start)
			}
			phrase = append(phrase, strings.ToLower(value))
		}
	}

	flushPhrase()
	tokens = append(tokens, Token{Type: TokenEOF})
	return tokens, nil
}
