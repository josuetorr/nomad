package node

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/gocolly/colly/v2"
	"github.com/josuetorr/nomad/internal/common"
	"github.com/josuetorr/nomad/internal/index"
	"github.com/josuetorr/nomad/internal/lexer"
)

const (
	cachedDir = "_cached"
	batchSize = 100
	numJobs   = 1000
)

const defaultSize = 1000

type (
	DocID = string
	Url   = string
	// NOTE: If we can find some clever way to organize document pages in a way to deliver better search
	// results, it might be a killer feature. Come back to it later
	// PageID string
)

type TokenizedDoc struct {
	DocID   DocID
	Url     Url
	content []string
}

type RawDoc struct {
	Url     Url
	Content string
}

type QueryResponse struct {
	Data []Url
}

type Query = string

type Node struct {
	kv *badger.DB
	i  index.Index
}

func NewNode(kv *badger.DB) Node {
	return Node{
		kv: kv,
		i:  index.Init(),
	}
}

// TODO: Respect robots.txt
func (n *Node) Crawl(startURL string) chan RawDoc {
	out := make(chan RawDoc, defaultSize)
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
			doc := RawDoc{
				Url:     url,
				Content: content,
			}
			out <- doc
		})
		c.Visit(startURL)
		c.Wait()
	}()
	return out
}

// NOTE: this will be skipped for now. No need for this complexity whilst prototyping
func (n *Node) SaveDocs() {}

func (n *Node) TokenizeDocs(docChan chan RawDoc) chan TokenizedDoc {
	out := make(chan TokenizedDoc, defaultSize)
	go func() {
		for d := range docChan {
			cid, err := common.HashCID([]byte(d.Content))
			if err != nil {
				slog.Error("Failed to create cid for: %s. Err: %s\n", string(d.Url), err)
				continue
			}
			l := lexer.NewLexer(d.Content)
			docID := cid.String()
			tdoc := TokenizedDoc{
				DocID:   docID,
				Url:     d.Url,
				content: l.Accumulate(),
			}
			out <- tdoc
		}
		close(out)
	}()
	return out
}

func (n *Node) AddDocs(tokenizeDocChan chan TokenizedDoc) chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		for tknDoc := range tokenizeDocChan {
			doc := index.CreateDoc(tknDoc.Url, tknDoc.content)
			n.i.AddDoc(tknDoc.DocID, doc)
		}
		done <- struct{}{}
	}()
	return done
}

// We lookup the td-idf table for the tokens in q
func (n *Node) Search(q Query) (*QueryResponse, error) {
	l := lexer.NewLexer(q)
	type doc struct {
		url string
		val float64
	}
	tfidfMap := make(map[string][]doc)
	for _, t := range l.Tokens() {
		t := string(t)
		docIDs := n.i.Terms[t]
		for _, dID := range docIDs {
			tf := n.i.TF(t, dID)
			idf := n.i.IDF(t)

			val := tf * idf
			doc := doc{
				url: n.i.Corpus[dID].Url,
				val: val,
			}
			tfidfMap[t] = append(tfidfMap[t], doc)
		}
		slices.SortFunc(tfidfMap[t], func(a, b doc) int {
			return int(a.val) - int(b.val)
		})
	}
	qres := QueryResponse{
		Data: []Url{},
	}
	for _, urls := range tfidfMap {
		for _, doc := range urls {
			qres.Data = append(qres.Data, doc.url)
		}
	}
	return &qres, nil
}
