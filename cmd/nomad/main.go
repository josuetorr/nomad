package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/badger/v4"
	"github.com/josuetorr/nomad/internal/node"
	"github.com/josuetorr/nomad/pkg/monad"
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
	println("Starting crawling")
	docs := n.Crawl(startURL)
	println("Starting tokenizint")
	tokenizedDocs := n.TokenizeDocs(docs)
	println("Starting filling corpus")
	done := n.AddDocs(tokenizedDocs)
	<-done
	println("Finished writing index TF...")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	println("Preparing for search query...")
	go func() {
		var q node.Query
		print("Enter your query: ")
		_, err := fmt.Scanln(&q)
		if err != nil {
			fmt.Printf("Failed to read query. Error: %s\n", err)
			return
		}
		res, err := n.Search(q)
		if err != nil {
			fmt.Printf("Failed to perform search: %s. Error: %s\n", string(q), err)
			return
		}
		println("Results: ")
		for _, url := range monad.Chopn(res.Data, 10) {
			fmt.Printf("	%s\n", url)
		}
	}()

	<-ctx.Done()
	println("Shutting down. Bye!")
}
