package spidey

import (
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly/v2"
	"github.com/josuetorr/nomad/internal/common"
)

const (
	cachedDir = "_cached"
)

type DocData struct {
	Url       string
	Content   string
	Indexable bool
}

type Spidey struct {
	kv common.KVStorer
}

func NewSpidey(kv common.KVStorer) Spidey {
	createDirIfNotExists(cachedDir)
	return Spidey{
		kv: kv,
	}
}

func (s Spidey) Crawl(startURL string, pc chan<- DocData) {
	c := colly.NewCollector(
		colly.Async(),
		colly.MaxDepth(2),
		colly.CacheDir(cachedDir),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
	c.OnHTML("a[href]", s.onHTML)
	c.OnResponse(s.onResponse(pc))
	c.Visit(startURL)
	c.Wait()
	close(pc)
}

func (s Spidey) onHTML(e *colly.HTMLElement) {
	link := e.Attr("href")
	e.Request.Visit(link)
}

func (s Spidey) onResponse(pc chan<- DocData) func(r *colly.Response) {
	return func(r *colly.Response) {
		url := r.Request.URL.String()
		k := common.DocKey(url)
		if s.kv.Exists(k) {
			pc <- DocData{Url: url, Indexable: false}
			return
		}

		compressed, err := common.Compress(r.Body)
		if err != nil {
			fmt.Printf("Failed to compress: %s. Error: %s\n", url, err)
			return
		}

		if err := s.kv.Put(k, compressed); err != nil {
			fmt.Printf("Failed to save: %s. Error: %s\n", url, err)
		}
		pc <- DocData{Url: url, Content: string(r.Body), Indexable: true}
	}
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create dir: %s. Error: %s\n", cachedDir, err)
		}
	}
}
