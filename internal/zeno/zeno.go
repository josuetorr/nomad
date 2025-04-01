package zeno

import (
	"fmt"
	"strings"
	"sync"

	"github.com/josuetorr/nomad/internal/common"
	"github.com/josuetorr/nomad/internal/lexer"
	"github.com/josuetorr/nomad/internal/spidey"
	"golang.org/x/net/html"
)

type (
	term     = string
	docID    = string
	docFreq  = map[docID]uint64
	termFreq = map[term]docFreq
)

type Zeno struct {
	kv common.KVStorer

	mu   *sync.Mutex
	tft  termFreq
	docN uint64
}

func NewZeno(kv common.KVStorer) Zeno {
	return Zeno{
		kv: kv,

		mu:   &sync.Mutex{},
		tft:  make(termFreq, 100),
		docN: 0,
	}
}

// IndexTF indexes the page received in cc channel by storing TermFrequence per Document.
// The calculation for tf-idf is calculate when a search query is received
func (z *Zeno) IndexTF(pc <-chan spidey.DocData) {
	for doc := range pc {
		z.docN++
		if !doc.Indexable {
			fmt.Printf("Skipping %s...\n", doc.Url)
			continue
		}

		root, err := html.Parse(strings.NewReader(doc.Content))
		if err != nil {
			fmt.Printf("Failed to parse: %s. Error: %s", doc.Url, err)
			continue
		}
		text := common.ExtractDocText(root)
		l := lexer.NewLexer(text)

		z.mu.Lock()
		for _, t := range l.Tokens() {
			t := string(t)
			if _, ok := z.tft[t]; !ok {
				z.tft[t] = make(docFreq, 100)
			} else {
				z.tft[t][doc.Url]++
			}
		}

		fmt.Printf("Indexing %s...\n", doc.Url)
		// TODO: use batching to write these entries
		for term, docF := range z.tft {
			i := 0
			for docID, f := range docF {
				if i == (len(docF) - 1) {
					z.kv.Put(common.TermKey(term), fmt.Appendf(nil, "%s:%d", docID, f))
				} else {
					z.kv.Put(common.TermKey(term), fmt.Appendf(nil, "%s:%d,", docID, f))
				}
				i++
			}
		}
		z.mu.Unlock()
	}
}

// IndexDF indexes documents by DocumentFrequency. This func should be called once IndexTF is done doing its job
// since it expects o.tf to not change
func (z *Zeno) IndexDF() {
	z.kv.Put(common.DocCountKey(), fmt.Appendf(nil, "%d", z.docN))
	for t, docF := range z.tft {
		k := common.DFKey(t)
		z.kv.Put(k, fmt.Appendf(nil, "%d", len(docF)))
	}
}
