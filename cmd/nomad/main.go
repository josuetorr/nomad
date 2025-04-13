package main

import (
	"fmt"

	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/node"
)

const startURL = "https://wikipedia.org/wiki/meme"

func main() {
	kv := db.NewKV("/tmp/badger/nodmad")
	n := node.NewNode(kv)
	docStage := n.Crawl(startURL)
	tokenizeStage := n.TokenizeDocs(docStage)
	done := n.IndexDF(tokenizeStage)
	// NOTE: for now we will omit saving docs. No need for prototyping. Docs are cached
	// TODO: index tf
	// TODO: index df

	<-done

	for docID, doc := range n.DfTable {
		fmt.Printf("doc ID: %s\n", docID)
		for t, tc := range doc {
			fmt.Printf("  term: %s, count: %d\n", t, tc)
		}
		println()
	}
}
