package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/josuetorr/nomad/internal/lexer"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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

type (
	Term  = string
	DocID = string

	TermFreq      = map[Term]int
	TermFreqIndex = map[DocID]TermFreq
)

func indexDoc(tfIndex TermFreqIndex) func(*colly.Response) {
	return func(r *colly.Response) {
		doc, err := html.Parse(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Fatalf("Could not parse document: %s: %s", r.Request.URL.String(), err)
		}

		url := r.Request.URL.String()
		if _, exists := tfIndex[url]; exists {
			fmt.Printf("Skipping: %s... already processed\n", url)
			return
		}

		fmt.Printf("Processing: %s...\n", url)
		content := extractDocText(doc)
		if content == "" {
			return
		}
		l := lexer.NewLexer(content)

		tc := make(TermFreq)
		for _, token := range l.Tokens() {
			t := string(token)
			if f, ok := tc[t]; !ok {
				tc[t] = 1
			} else {
				tc[t] = f + 1
			}
		}

		tfIndex[url] = tc
	}
}

func main() {
	const cachedDir = "_cached"
	createDirIfNotExists(cachedDir)

	const startingUrl = "https://wikipedia.org/wiki/meme"

	tfIndex := make(TermFreqIndex)
	c := colly.NewCollector(colly.MaxDepth(2), colly.CacheDir(cachedDir))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	c.OnResponse(indexDoc(tfIndex))
	c.Visit(startingUrl)
}
