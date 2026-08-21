package parser

import (
	"fmt"

	ast "github.com/GabrielMoody/Cariin/internal/search/AST"
	"github.com/GabrielMoody/Cariin/internal/search/token"
)

type parser struct {
	tokens   []token.Token
	position int
}

func Parse(input string) (ast.Query, error) {
	tokens, err := token.Lex(input)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 1 {
		return nil, fmt.Errorf("query is empty")
	}

	p := parser{tokens: tokens}
	query, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.current().Type != token.TokenEOF {
		return nil, fmt.Errorf("unexpected token %q", p.current().Value)
	}
	return query, nil
}

func (p *parser) parseOr() (ast.Query, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.match(token.TokenOR) {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = ast.OrQuery{Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (ast.Query, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.match(token.TokenAND) {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = ast.AndQuery{Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseUnary() (ast.Query, error) {
	if p.match(token.TokenNOT) {
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return ast.NotQuery{Query: operand}, nil
	}
	if p.match(token.TokenLeftParen) {
		query, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if !p.match(token.TokenRightParen) {
			return nil, fmt.Errorf("expected closing parenthesis")
		}
		return query, nil
	}

	current := p.current()
	if current.Type != token.TokenTerm {
		return nil, fmt.Errorf("expected term, got %q", current.Value)
	}
	p.position++
	return ast.TermQuery{Term: current.Value}, nil
}

func (p *parser) current() token.Token {
	return p.tokens[p.position]
}

func (p *parser) match(kind token.TokenType) bool {
	if p.current().Type != kind {
		return false
	}
	p.position++
	return true
}
