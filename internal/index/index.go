package index

import "math"

type (
	DocID = string
	Url   = string
	Term  = string
	Doc   struct {
		Url Url
		TF  map[Term]uint64
	}
	Index struct {
		corpus map[DocID]*Doc
		terms  map[Term][]DocID
	}
)

const defaultSize = 1000

func Init() Index {
	return Index{
		corpus: make(map[DocID]*Doc, defaultSize),
		terms:  make(map[Term][]DocID, defaultSize),
	}
}

func (i *Index) AddDoc(dID DocID, d *Doc) {
	i.corpus[dID] = d
	for t := range d.TF {
		if _, ok := i.terms[t]; !ok {
			i.terms[t] = make([]DocID, defaultSize)
		}
		i.terms[t] = append(i.terms[t], dID)
	}
}

func (i *Index) TF(t Term, dID DocID) float64 {
	doc := i.corpus[dID]
	termCount := len(doc.TF)
	return math.Log(float64(1 + (doc.TF[t] / uint64(termCount))))
}

func (i *Index) IDF(t Term) float64 {
	corpusSize := len(i.corpus)
	docs := i.terms[t]
	return math.Log(float64(corpusSize / len(docs)))
}
