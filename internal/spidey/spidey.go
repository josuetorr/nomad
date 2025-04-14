package spidey

import (
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/josuetorr/nomad/pkg/pipeline"
)

const (
	cachedDir = "_cached"
)

type DocData struct {
	Url     string
	Content string
}

type Spidey struct{}

func NewSpidey() Spidey {
	createDirIfNotExists(cachedDir)
	return Spidey{}
}

func (s Spidey) Crawl(startURL string) pipeline.Stage[DocData] {
	return func() chan DocData {
		out := make(chan DocData, 1000)
		go func() {
			defer close(out)
			c := colly.NewCollector(
				colly.Async(),
				colly.MaxDepth(3),
				colly.CacheDir(cachedDir),
			)
			c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
			c.OnHTML("a[href]", s.onHTML)
			c.OnResponse(s.onResponse(out))
			c.Visit(startURL)
			c.Wait()
		}()
		return out
	}
}

func (s Spidey) onHTML(e *colly.HTMLElement) {
}

func (s Spidey) onResponse(c chan DocData) func(r *colly.Response) {
	return func(r *colly.Response) {
		// TODO: handle other types of documents
		if !strings.Contains(r.Headers.Get("Content-Type"), "text/html") {
			return
		}
		url := r.Request.URL.String()
		content := string(r.Body)
		doc := DocData{
			Url:     url,
			Content: content,
		}
		c <- doc
	}
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create dir: %s. Error: %s\n", cachedDir, err)
		}
	}
}
