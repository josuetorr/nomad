package index

import (
	"math"
	"sync"
)

type (
	DocID = string
	Url   = string
	Term  = string
	Doc   struct {
		Url Url
		TF  map[Term]uint64
	}
	Index struct {
		corpusLock *sync.RWMutex
		Corpus     map[DocID]*Doc

		termsLock *sync.RWMutex
		Terms     map[Term][]DocID
	}
)

const defaultSize = 1000

func Init() Index {
	return Index{
		corpusLock: &sync.RWMutex{},
		Corpus:     make(map[DocID]*Doc, defaultSize),

		termsLock: &sync.RWMutex{},
		Terms:     make(map[Term][]DocID, defaultSize),
	}
}

func (i *Index) AddDoc(dID DocID, d *Doc) {
	i.corpusLock.Lock()
	i.Corpus[dID] = d
	i.corpusLock.Unlock()

	i.termsLock.Lock()
	for t := range d.TF {
		if _, ok := i.Terms[t]; !ok {
			i.Terms[t] = []DocID{}
		}
		i.Terms[t] = append(i.Terms[t], dID)
	}
	i.termsLock.Unlock()
}

func (i *Index) TF(t Term, dID DocID) float64 {
	defer i.corpusLock.RUnlock()
	i.corpusLock.RLock()
	doc := i.Corpus[dID]
	if doc == nil {
		return 0
	}
	return math.Log(float64(1 + doc.TF[t]))
}

func (i *Index) IDF(t Term) float64 {
	defer i.termsLock.RUnlock()
	i.termsLock.RLock()
	corpusSize := len(i.Corpus)
	docs := i.Terms[t]
	return math.Log((float64(corpusSize+1) / float64(len(docs)+1)))
}

func (i *Index) CorpusSize() int {
	return len(i.Corpus)
}
