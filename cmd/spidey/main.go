package main

import (
	"fmt"
	"log"
	"os"
	"sort"
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
	TermFreq = map[string]float32
	Index    = map[string]TermFreq
)

func indexDocument(index Index) func(*colly.Response) {
	return func(r *colly.Response) {
		doc, err := html.Parse(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Fatalf("Could not parse document: %s: %s", r.Request.URL.String(), err)
		}

		url := r.Request.URL.String()
		if _, exists := index[url]; exists {
			fmt.Printf("Skipping: %s... already indexed\n", url)
			return
		}

		fmt.Printf("Indexing: %s...\n", url)
		content := extractDocText(doc)
		l := lexer.NewLexer(content)

		tf := make(TermFreq)
		for _, token := range l.Tokens() {
			t := string(token)
			if f, ok := tf[t]; !ok {
				tf[t] = 1
			} else {
				tf[t] = f + 1
			}
		}

		for t, f := range tf {
			tf[t] = f / float32(len(tf))
		}
		index[url] = tf
	}
}

func main() {
	const cachedDir = "_cached"
	createDirIfNotExists(cachedDir)

	startingUrl := "https://wikipedia.org/wiki/meme"

	index := make(Index)
	c := colly.NewCollector(colly.MaxDepth(2), colly.CacheDir(cachedDir))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	c.OnResponse(indexDocument(index))
	c.Visit(startingUrl)

	for url, tf := range index {
		const topN = 10
		fmt.Println("-------------")
		fmt.Printf("tf-idf for: %s\n", url)
		fmt.Printf("top %d terms\n", topN)

		keys := make([]string, 0, len(tf))
		for k := range tf {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return tf[keys[i]] > tf[keys[j]]
		})

		for _, k := range keys[0:topN] {
			fmt.Printf("term: %s, tf: %f\n", k, tf[k])
		}
	}
}
