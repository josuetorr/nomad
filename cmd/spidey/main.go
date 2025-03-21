package main

import (
	"fmt"
	"iter"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/gocolly/colly"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Token = []rune

type Lexer struct {
	content []rune
}

func NewLexer(s string) *Lexer {
	return &Lexer{content: []rune(s)}
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

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func extractDocText(root *html.Node) string {
	var textBuilder strings.Builder
	ignoredTags := map[atom.Atom]bool{
		atom.Style:  true,
		atom.Script: true,
	}

	for n := range root.Descendants() {
		if n.Type != html.TextNode {
			continue
		}

		parent := n.Parent
		if parent != nil && parent.Type == html.ElementNode {
			if ignoredTags[parent.DataAtom] {
				continue
			}
		}

		text := strings.TrimSpace(n.Data)
		if text != "" {
			if textBuilder.Len() > 0 {
				textBuilder.WriteString(" ")
			}
			textBuilder.WriteString(text)
		}
	}

	return textBuilder.String()
}

func parseResponse(r *colly.Response) {
	doc, err := html.Parse(strings.NewReader(string(r.Body)))
	if err != nil {
		log.Fatalf("Could not parse document: %s: %s", r.Request.URL.String(), err)
	}

	content := extractDocText(doc)
	l := NewLexer(content)

	for _, t := range l.Tokens() {
		fmt.Printf("%+v\n", string(t))
	}
}

func main() {
	cachedDir := "_cached"
	createDirIfNotExists(cachedDir)

	startingUrl := "https://wikipedia.org/wiki/meme"

	c := colly.NewCollector(colly.MaxDepth(1), colly.CacheDir(cachedDir))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	c.OnResponse(parseResponse)
	c.Visit(startingUrl)
}
