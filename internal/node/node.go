package node

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/dgraph-io/badger/v4"
	"github.com/gocolly/colly/v2"
	"github.com/josuetorr/nomad/internal/common"
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
	kv *badger.DB
}

func NewNode(kv *badger.DB) Node {
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

func (n *Node) WriteIndexDF(dtfChan <-chan DocTermFreq) {
	doneWritingDocs := make(chan struct{})
	var docCount atomic.Uint64
	go func() {
		for dtf := range dtfChan {
			k := common.DocKey(string(dtf.DocID))
			v := ""
			for t, f := range dtf.TF {
				v = fmt.Sprintf("%s%s%s%d,", v, t, common.KeySep, f)
			}
			fmt.Printf("DF indexing doc: %s...\n", dtf.DocID)
			if err := n.kv.Update(func(txn *badger.Txn) error {
				return txn.Set([]byte(k), []byte(v))
			}); err == nil {
				docCount.Add(1)
			}
		}
		doneWritingDocs <- struct{}{}
	}()
	<-doneWritingDocs
	println("attempting to update doc count")
	n.kv.Update(func(txn *badger.Txn) error {
		k := []byte(common.DocCountKey())
		item, err := txn.Get(k)
		if errors.Is(err, badger.ErrKeyNotFound) {
			v := common.Uint64ToBytes(0)
			txn.Set(k, v)
		}
		if err != nil {
			return err
		}
		bytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		savedv, err := common.BytesToUint64(bytes)
		if err != nil {
			return err
		}
		dc := docCount.Load()
		txn.Set(k, common.Uint64ToBytes(dc+savedv))
		return nil
	})
}

func (n *Node) WriteIndexTF() error {
	tf := make(TFIndex, defaultSize)
	addDocDFEntry := func(key, val []byte) error {
		k := string(key)
		parts := common.KeyParts(k)
		if len(parts) != 2 {
			return fmt.Errorf("doc key must only have 2 parts: %s\n", k)
		}
		docID := DocID(parts[1])
		dfIndexRow := string(val)
		for tfValue := range strings.SplitSeq(dfIndexRow, ",") {
			tfParts := strings.Split(tfValue, common.KeySep)
			if len(tfParts) != 2 {
				// NOTE: for now ignore values that have invalid formats
				// will try to fix using protobuf for serializing
				return nil
			}
			t := tfParts[0]
			fStr := tfParts[1]
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
	}
	n.kv.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		p := []byte(common.DocCountKey())
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				return addDocDFEntry(key, val)
			})
			if err != nil {
				fmt.Printf("Failed to get value for: %s. err: %s\n", string(key), err)
			}
		}
		return nil
	})
	wb := n.kv.NewWriteBatch()
	defer wb.Cancel()
	for t, df := range tf {
		v := ""
		for docID, dc := range df {
			v = fmt.Sprintf("%s%s%s%d,", v, docID, common.KeySep, dc)
		}
		fmt.Printf("TF Indexing term: %s...\n", t)
		if err := wb.Set([]byte(common.TermKey(t)), []byte(v)); err != nil {
			return err
		}
	}
	return wb.Flush()
}

// We lookup the td-idf table for the tokens in q
func (n *Node) Search(q string) error {
	// l := lexer.NewLexer(q)
	// for _, t := range l.Tokens() {
	// 	t := string(t)
	// }
	// return nil
	panic("todo search")
}
