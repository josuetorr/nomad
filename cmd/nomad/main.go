package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/josuetorr/nomad/internal/common"
	"github.com/josuetorr/nomad/internal/node"
)

const (
	startURL    = "https://wikipedia.org/wiki/meme"
	nomadKvPath = "/tmp/badger/nomad"
)

func initKV(path string) *badger.DB {
	kv, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		log.Fatalf("Failed to open kv store\n%s", err)
	}
	return kv
}

func main2() {
	kv := initKV(nomadKvPath)
	n := node.NewNode(kv)
	docs := n.Crawl(startURL)
	tokens := n.TokenizeDocs(docs)
	doctfs := n.DFIndex(tokens)
	n.WriteIndexDF(doctfs)
	println("Starting to write index TF...")
	n.WriteIndexTF()

	// TODO: try to handle a search query
	// q := "what is a meme?"
	// n.Search(q)
}

func main() {
	kv := initKV(nomadKvPath)
	kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(common.DocCountKey()))
		if errors.Is(err, badger.ErrKeyNotFound) {
			fmt.Printf("Could not find doc count entry")
		}
		bytes, err := item.ValueCopy(nil)
		if err != nil {
			log.Fatalf("Failed to copy doc count value")
		}
		dc, err := common.BytesToUint64(bytes)
		if err != nil {
			log.Fatalf("Failed to convert bytes into uint64")
		}
		fmt.Printf("doc_count: %d\n", dc)
		return nil
	})
}
