package zeno

import (
	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/spidey"
)

type (
	term     = string
	docID    = string
	docFreq  = map[docID]uint64
	termFreq = map[term]docFreq
)

type Zeno struct{}

const mapSize = 500

func NewZeno(kv db.KVStorer) Zeno {
	return Zeno{}
}

// IndexTF indexes the page received in cc channel by storing TermFrequence per Document.
// The calculation for tf-idf is calculate when a search query is received
func (z *Zeno) IndexTF(pc <-chan spidey.DocData) {
	for doc := range pc {
	}
}

// IndexDF indexes documents by DocumentFrequency. This func should be called once IndexTF is done doing its job
// since it expects o.tf to not change
func (z *Zeno) IndexDF() {
}
