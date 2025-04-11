package node

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/josuetorr/nomad/internal/common"
	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/lexer"
)

const (
	cachedDir = "_cached"
	batchSize = 100
	numJobs   = 1000
)

type doc struct {
	Url     string
	Content string
}

type tokenizedDoc struct {
	url    string
	tokens []string
}

type (
	termFreq   = map[string]uint64
	indexedDoc struct {
		url string
		tf  termFreq
	}
)

// TODO: abstract away the process of flushing batch to chan when batch is full
// and when all jobs have been process but the batch still has some data
type Node struct {
	kv db.KVStorer

	docBatch []doc
	DocJobs  chan []doc

	tokenizedDocBatch []tokenizedDoc
	TokenizedDocJobs  chan []tokenizedDoc

	indexedDocBatch []indexedDoc
	indexedDocJobs  chan []indexedDoc
}

func Init(kv db.KVStorer) Node {
	return Node{
		kv: kv,

		docBatch: make([]doc, 0, batchSize),
		DocJobs:  make(chan []doc, numJobs),

		tokenizedDocBatch: make([]tokenizedDoc, 0, batchSize),
		TokenizedDocJobs:  make(chan []tokenizedDoc, numJobs),

		indexedDocBatch: make([]indexedDoc, 0, batchSize),
		indexedDocJobs:  make(chan []indexedDoc, numJobs),
	}
}

func (n *Node) Crawl(startURL string) {
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
		if len(n.docBatch) == batchSize {
			n.DocJobs <- n.docBatch
			n.docBatch = make([]doc, 0, batchSize)
		} else {
			n.docBatch = append(n.docBatch, d)
		}
	})
	c.Visit(startURL)
	c.Wait()
	if len(n.docBatch) > 0 {
		n.DocJobs <- n.docBatch
		n.docBatch = nil
	}
	close(n.DocJobs)
}

func (n *Node) SaveDocs() {
	for dj := range n.DocJobs {
		n.kv.BatchWrite(func(w db.KVWriter) {
			for _, d := range dj {
				k := common.DocKey(d.Url)
				compressed, _ := common.Compress([]byte(d.Content))
				if err := w.Set([]byte(k), compressed); err != nil {
					fmt.Printf("Failed to save %s. err: %s\n", d.Url, err)
				}
			}
		})
	}
}

func (n *Node) TokenizeDocs() {
	for dj := range n.DocJobs {
		for _, d := range dj {
			l := lexer.NewLexer(d.Content)
			td := tokenizedDoc{
				url: d.Url,
			}
			for _, t := range l.Tokens() {
				t := string(t)
				td.tokens = append(td.tokens, t)
			}
			if len(n.tokenizedDocBatch) == batchSize {
				n.TokenizedDocJobs <- n.tokenizedDocBatch
				n.tokenizedDocBatch = make([]tokenizedDoc, 0, batchSize)
			} else {
				n.tokenizedDocBatch = append(n.tokenizedDocBatch, td)
			}
		}
	}
	if len(n.tokenizedDocBatch) > 0 {
		n.TokenizedDocJobs <- n.tokenizedDocBatch
		n.tokenizedDocBatch = nil
	}
}

func (n *Node) IndexTF() {
	for tdj := range n.TokenizedDocJobs {
		for _, td := range tdj {
			for _, t := range td.tokens {
				tf := make(termFreq, len(td.tokens))
				tf[t]++
				idxd := indexedDoc{
					url: td.url,
					tf:  tf,
				}

				if len(n.tokenizedDocBatch) == batchSize {
					n.indexedDocJobs <- n.indexedDocBatch
					n.tokenizedDocBatch = make([]tokenizedDoc, 0, batchSize)
				} else {
					n.indexedDocBatch = append(n.indexedDocBatch, idxd)
				}
			}
			if len(n.tokenizedDocBatch) > 0 {
				n.indexedDocJobs <- n.indexedDocBatch
				n.tokenizedDocBatch = nil
			}
		}
	}
}

func (n *Node) SaveIndexTF() {
	for idj := range n.indexedDocJobs {
		for _, idxd := range idj {
			for t := range idxd.tf {
				// TODO: figure out how to write to db
			}
		}
	}
}
func (n *Node) indexDF()     {}
func (n *Node) saveIndexDF() {}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create dir: %s. Error: %s\n", dir, err)
		}
	}
}
