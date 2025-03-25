package main

import (
	"fmt"
	"log"
	"math"
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

func tf(t Term, tf TermFreq) float64 {
	if _, ok := tf[t]; !ok {
		return 0
	}

	sum := 0
	for _, f := range tf {
		sum += f
	}

	return float64(tf[t]) / float64(sum)
}

func idf(term Term, docs TermFreqIndex) float64 {
	docN := len(docs)

	termInDocCount := 0
	for _, tf := range docs {
		_, ok := tf[term]
		if !ok {
			continue
		}
		termInDocCount += 1
	}

	return math.Log((float64(docN) + 1) / (float64(termInDocCount) + 1))
}

func main() {
	args := os.Args[1:]

	// args parsing
	if len(args) != 1 {
		log.Fatal("Invalid args. Must only provide query")
	}

	const cachedDir = "_cached"
	createDirIfNotExists(cachedDir)

	const startingUrl = "https://wikipedia.org/wiki/meme"

	// indexing
	tfIndex := make(TermFreqIndex)
	c := colly.NewCollector(colly.MaxDepth(2), colly.CacheDir(cachedDir))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	c.OnResponse(indexDoc(tfIndex))
	c.Visit(startingUrl)

	// calculating td-idf
	q := args[0]
	l := lexer.NewLexer(q)
	tokens := []string{}
	for _, t := range l.Tokens() {
		tokens = append(tokens, string(t))
	}
	for docID, termfreq := range tfIndex {
		fmt.Printf("%s\n", docID)
		for _, t := range tokens {
			t := string(t)
			tf := tf(t, termfreq)
			idf := idf(t, tfIndex)
			fmt.Printf("  %s => tf: %f, idf: %f, tfidf: %f\n", t, tf, idf, tf*idf)
		}
		println()
	}
}
