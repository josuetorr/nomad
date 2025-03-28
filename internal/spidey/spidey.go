package spidey

import (
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
	"github.com/josuetorr/nomad/internal/common"
)

const (
	cachedDir = "_cached"
)

type CrawledPage struct {
	Url     string
	Content string
}

type Spidey struct {
	store common.Storer
}

func NewSpidey(store common.Storer) Spidey {
	createDirIfNotExists(cachedDir)
	return Spidey{
		store: store,
	}
}

func (s Spidey) Crawl(entryPoint string, cc chan<- CrawledPage) {
	// NOTE: change depth once crawler / indexer communication has been established
	c := colly.NewCollector(colly.MaxDepth(1), colly.CacheDir(cachedDir))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	c.OnScraped(s.onScrapped(cc))
	c.Visit(entryPoint)
}

func (s Spidey) onScrapped(cc chan<- CrawledPage) func(r *colly.Response) {
	return func(r *colly.Response) {
		url := r.Request.URL.String()
		k := common.DocKey(url)
		ok := s.store.Exists(k)
		if ok {
			fmt.Printf("Skipping %s... already saved\n", url)
			return
		}

		cc <- CrawledPage{Url: url, Content: string(r.Body)}

		compressed, err := common.Compress(r.Body)
		if err != nil {
			log.Fatalf("Failed to compress: %s. Error :%s", url, err)
		}
		fmt.Printf("Saving %s...\n", url)
		if err := s.store.Put(k, compressed); err != nil {
			log.Fatalf("Failed to store doc: %s. Error: %s", url, err)
		}
	}
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create dir: %s. Error: %s", cachedDir, err)
		}
	}
}
