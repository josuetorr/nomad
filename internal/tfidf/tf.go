package tfidf

type (
	url      = string
	TermFreq map[url]int
)

func (tf TermFreq) Tf(term string) int {
	f, exists := tf[term]
	if !exists {
		return 0
	}
	return f / len(tf)
}
