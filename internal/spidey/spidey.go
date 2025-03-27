package spidey

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
	"github.com/josuetorr/nomad/internal/common"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// NOTE: might move it somewhere else
func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	defer w.Close()
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NOTE: might move it somewhere else
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

const (
	cachedDir = "_cached"
)

type Spidey struct {
	store common.Storer
}

func NewSpidey(store common.Storer) Spidey {
	return Spidey{
		store: store,
	}
}

func (s Spidey) Crawl(startingUrl string) {
	c := colly.NewCollector(colly.MaxDepth(2), colly.CacheDir(cachedDir))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	c.OnScraped(s.onScrapped)
	c.Visit(startingUrl)
}

func (s Spidey) onScrapped(r *colly.Response) {
	url := r.Request.URL.String()
	k := common.PageKey + url
	ok := s.store.Exists(k)
	if ok {
		fmt.Printf("Skipping %s... already saved\n", url)
		return
	}

	doc, err := html.Parse(strings.NewReader(string(r.Body)))
	if err != nil {
		log.Fatalf("Could not parse document: %s. Error: %s", r.Request.URL.String(), err)
	}

	content := extractDocText(doc)
	if content == "" {
		return
	}

	compressed, err := compress([]byte(content))
	fmt.Printf("Saving %s...\n", url)
	if err := s.store.Put(k, compressed); err != nil {
		log.Fatalf("Failed to store doc: %s. Error: %s", url, err)
	}
}
