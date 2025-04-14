package main

import (
	"fmt"

	"github.com/josuetorr/nomad/internal/common"
	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/node"
)

const startURL = "https://wikipedia.org/wiki/meme"

func main() {
	kv := db.NewKV("/tmp/badger/nomad")
	n := node.NewNode(kv)
	docs := n.Crawl(startURL)
	tokens := n.TokenizeDocs(docs)
	doctfs := n.DFIndex(tokens)
	done := n.WriteIndexDF(doctfs)

	err := <-done
	if err != nil {
		fmt.Printf("Err: %s\n", err)
	} else {
		n.WriteIndexTF()
	}
}

func main2() {
	kv := db.NewKV("/tmp/badger/nomad")
	bytes, _ := kv.Get(common.DocKey(startURL))
	fmt.Printf("%s\n", string(bytes))
}
