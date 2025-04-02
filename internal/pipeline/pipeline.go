package pipeline

import (
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

const (
	cachedDir = "_cached"
	batchSize = 100
	jobsNum   = 1000
)

type doc struct {
	Url     string
	Content string
}

type tokenizedDoc struct {
	url    string
	tokens []string
}

type Pipeline struct {
	docBatch []doc
	DocJobs  chan []doc

	tokenizedDocBatch []tokenizedDoc
	TokenizedDocJobs  chan []tokenizedDoc
}

func Init() Pipeline {
	return Pipeline{
		docBatch: make([]doc, 0, batchSize),
		DocJobs:  make(chan []doc, jobsNum),

		tokenizedDocBatch: make([]tokenizedDoc, 0, batchSize),
		TokenizedDocJobs:  make(chan []tokenizedDoc, jobsNum),
	}
}

func (p *Pipeline) Crawl(startURL string) {
	createDirIfNotExists(cachedDir)
	c := colly.NewCollector(
		colly.Async(),
		colly.MaxDepth(2),
		colly.CacheDir(cachedDir),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	c.OnResponse(func(r *colly.Response) {
		if !strings.Contains(r.Headers.Get("Content-Type"), "text/html") {
			return
		}
		url := r.Request.URL.String()
		d := doc{
			Url:     url,
			Content: string(r.Body),
		}
		if len(p.docBatch) == batchSize {
			p.DocJobs <- p.docBatch
			p.docBatch = make([]doc, 0, batchSize)
		} else {
			p.docBatch = append(p.docBatch, d)
		}
	})
	c.Visit(startURL)
	c.Wait()
	if len(p.docBatch) > 0 {
		p.DocJobs <- p.docBatch
		p.docBatch = nil
	}
	close(p.DocJobs)
}

func (p *Pipeline) saveDocs()     {}
func (p *Pipeline) tokenizeDocs() {}
func (p *Pipeline) indexTF()      {}
func (p *Pipeline) saveIndexTF()  {}
func (p *Pipeline) indexDF()      {}
func (p *Pipeline) saveIndexDF()  {}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create dir: %s. Error: %s\n", cachedDir, err)
		}
	}
}
