package node

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/lexer"
)

const (
	cachedDir = "_cached"
	batchSize = 100
	numJobs   = 1000
)

const defaultSize = 1000

type DocId string

type TokenizedDoc struct {
	DocID   DocId
	content []string
}

type DfTable map[DocId]map[string]uint64

type Doc struct {
	DocId   DocId
	Content string
}

type Node struct {
	kv db.KVStorer

	rw      *sync.Mutex
	DfTable DfTable
}

func NewNode(kv db.KVStorer) Node {
	return Node{
		kv:      kv,
		rw:      &sync.Mutex{},
		DfTable: make(DfTable, defaultSize),
	}
}

func (n *Node) Crawl(startURL string) chan Doc {
	out := make(chan Doc, defaultSize)
	go func() {
		defer close(out)
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
			// TODO: handle other types of documents
			if !strings.Contains(r.Headers.Get("Content-Type"), "text/html") {
				return
			}
			url := r.Request.URL.String()
			content := string(r.Body)
			doc := Doc{
				DocId:   DocId(url),
				Content: content,
			}
			out <- doc
		})
		c.Visit(startURL)
		c.Wait()
	}()
	return out
}

func (n *Node) SaveDocs() {
	// NOTE: this will be skipped for now. No need for this complexity whilst prototyping
}

func (n *Node) TokenizeDocs(docChan chan Doc) chan TokenizedDoc {
	out := make(chan TokenizedDoc, defaultSize)
	go func() {
		for d := range docChan {
			tdoc := TokenizedDoc{}
			l := lexer.NewLexer(d.Content)
			for _, t := range l.Tokens() {
				tdoc.DocID = d.DocId
				tdoc.content = append(tdoc.content, string(t))
			}
			out <- tdoc
		}
		close(out)
	}()
	return out
}

func (n *Node) IndexDF(tokenizeDocChan chan TokenizedDoc) chan struct{} {
	done := make(chan struct{})
	go func() {
		n.rw.Lock()
		for tdoc := range tokenizeDocChan {
			fmt.Printf("df indexing: %s\n", tdoc.DocID)
			if _, ok := n.DfTable[tdoc.DocID]; !ok {
				n.DfTable[tdoc.DocID] = make(map[string]uint64, defaultSize)
			}
			for _, t := range tdoc.content {
				n.DfTable[tdoc.DocID][t]++
			}
		}
		n.rw.Unlock()
		close(done)
	}()
	return done
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create dir: %s. Error: %s\n", dir, err)
		}
	}
}
