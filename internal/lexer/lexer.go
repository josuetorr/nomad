package lexer

import (
	"iter"
	"unicode"
)

type Token = []rune

type Lexer struct {
	content Token
}

func NewLexer(s string) *Lexer {
	return &Lexer{
		content: Token(s),
	}
}

func (l *Lexer) trimLeft() {
	for len(l.content) > 0 && unicode.IsSpace(l.content[0]) {
		l.content = l.content[1:]
	}
}

func (l *Lexer) chop(n int) Token {
	token := l.content[0:n]
	l.content = l.content[n:]
	return token
}

func (l *Lexer) chopWhile(p func(rune) bool) Token {
	n := 0
	for len(l.content) > n && p(l.content[n]) {
		n++
	}
	return l.chop(n)
}

func (l *Lexer) nextToken() *Token {
	l.trimLeft()
	if len(l.content) == 0 {
		return nil
	}

	p := unicode.IsLetter
	if p(l.content[0]) {
		token := l.chopWhile(p)
		return &token
	}

	p = unicode.IsNumber
	if p(l.content[0]) {
		token := l.chopWhile(p)
		return &token
	}

	token := l.chop(1)
	return &token
}

func (l *Lexer) Accumulate() []string {
	tokens := []string{}
	for _, t := range l.Tokens() {
		tokens = append(tokens, string(t))
	}
	return tokens
}

func (l *Lexer) Tokens() iter.Seq2[int, Token] {
	return func(yield func(int, Token) bool) {
		t := l.nextToken()
		for i := 0; t != nil; i++ {
			if !yield(i, *t) {
				return
			}
			t = l.nextToken()
		}
	}
}
