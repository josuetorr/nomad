package node

import (
	"fmt"
	"strconv"
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

const defaultSize = 1000

type (
	DocID string
	// NOTE: If we can find some clever way to organize document pages in a way to deliver better search
	// results, it might be a killer feature. Come back to it later
	// PageID string
)

type TokenizedDoc struct {
	DocID   DocID
	content []string
}

type Doc struct {
	DocId   DocID
	Content string
}

type (
	TermFreq    map[string]uint64
	DocTermFreq struct {
		DocID DocID
		TF    TermFreq
	}
	DocFreq map[DocID]uint64
	TFIndex map[string]DocFreq
)

// NOTE: seems like a way to write more efficiently. Piggybacking on the PageID idea
// just keeping the idea here
// type DFIndex map[DocID]TermFreq
// type DFIndex     map[PageID]DFIndex

type Node struct {
	kv db.KVStorer
}

func NewNode(kv db.KVStorer) Node {
	return Node{
		kv: kv,
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
				DocId:   DocID(url),
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

func (n *Node) DFIndex(tokenizeDocChan chan TokenizedDoc) chan DocTermFreq {
	out := make(chan DocTermFreq, defaultSize)
	go func() {
		for tdoc := range tokenizeDocChan {
			dtf := DocTermFreq{
				DocID: tdoc.DocID,
				TF:    make(TermFreq, defaultSize),
			}
			for _, t := range tdoc.content {
				dtf.TF[t]++
			}
			out <- dtf
		}
		close(out)
	}()
	return out
}

func (n *Node) WriteIndexDF(dtfChan <-chan DocTermFreq) chan error {
	done := make(chan error)
	go func() {
		for dtf := range dtfChan {
			k := common.DocKey(string(dtf.DocID))
			v := ""
			for t, f := range dtf.TF {
				v = fmt.Sprintf("%s%s%s%d,", v, t, common.KeySep, f)
			}
			fmt.Printf("indexing %s...\n", dtf.DocID)
			if err := n.kv.Put(k, []byte(v)); err != nil {
				done <- err
			}
		}
		done <- nil
	}()
	return done
}

func (n *Node) WriteIndexTF() error {
	tf := make(TFIndex, defaultSize)
	err := n.kv.IteratePrefix(common.DocKey(), func(key, val []byte) error {
		k := string(key)
		parts := common.KeyParts(k)
		if len(parts) != 2 {
			return fmt.Errorf("doc key must only have 2 parts: %s ()\n", k)
		}
		docID := DocID(parts[1])
		println(docID)

		v := string(val)
		for tfValue := range strings.SplitSeq(v, ",") {
			tfStr := strings.Split(tfValue, common.KeySep)
			if len(tfStr) != 2 {
				return fmt.Errorf("term value must only have 2 parts: %s\n", k)
			}
			t := tfStr[0]
			fStr := tfStr[1]
			_, ok := tf[t]
			if !ok {
				tf[t] = make(DocFreq, defaultSize)
			}
			f, err := strconv.Atoi(fStr)
			if err != nil {
				return err
			}
			tf[t][docID] += uint64(f)
		}

		return nil
	})
	fmt.Printf("%+v\n", tf)
	return err
}
