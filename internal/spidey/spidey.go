package spidey

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
	"github.com/josuetorr/nomad/internal/common"
	"golang.org/x/net/html"
)

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

	content := common.ExtractDocText(doc)
	if content == "" {
		return
	}

	compressed, err := common.Compress([]byte(content))
	fmt.Printf("Saving %s...\n", url)
	if err := s.store.Put(k, compressed); err != nil {
		log.Fatalf("Failed to store doc: %s. Error: %s", url, err)
	}
}
