package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"

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

func main() {
	// for now this needs to run before we can perform a query.
	kv := initKV(nomadKvPath)
	n := node.NewNode(kv)
	docs := n.Crawl(startURL)
	tokens := n.TokenizeDocs(docs)
	doctfs := n.DFIndex(tokens)
	n.WriteIndexDF(doctfs)
	println("Starting to write index TF...")
	if err := n.WriteIndexTF(); err != nil {
		fmt.Printf("Failed to write index TF")
		return
	}
	println("Finished writing index TF...")

	qs := make(chan node.Query)
	errs := make(chan error)
	cleanup := func() {
		defer close(qs)
		defer close(errs)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	println("Preparing for search query...")
	for {
		select {
		case q := <-qs:
			go func(q node.Query, errs chan<- error) {
				if err := n.Search(q); err != nil {
					errs <- err
				}
			}(q, errs)
		case err := <-errs:
			defer cleanup()
			fmt.Printf("Err: %s\n", err)
			fmt.Printf("Closing query channel...\n")
			return
		case <-ctx.Done():
			defer cleanup()
			fmt.Printf("Got signal interrupt")
			return
		default:
			var q string
			fmt.Print("Enter your query: ")
			fmt.Scanf("%s", &q)
			go func(q node.Query) {
				qs <- q
			}(node.Query(q))

		}
	}
}

// using for prototyping
func main2() {
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
