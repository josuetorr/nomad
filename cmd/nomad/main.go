package main

import (
	"sync"

	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/node"
)

const startURL = "https://wikipedia.org/wiki/meme"

func main() {
	kv := db.NewKV("/tmp/badger/nomad")
	n := node.Init(kv)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.Crawl(startURL)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.SaveDocs()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.TokenizeDocs()
	}()
	wg.Wait()
	println("done")
}
