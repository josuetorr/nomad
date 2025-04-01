package main

import (
	"sync"

	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/spidey"
	"github.com/josuetorr/nomad/internal/zeno"
)

const startURL = "https://wikipedia.org/wiki/meme"

func main() {
	kv := db.NewKV("/tmp/badger/nomad")
	pc := make(chan spidey.DocData, 100)
	spidey := spidey.NewSpidey(kv)
	zeno := zeno.NewZeno(kv)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		spidey.Crawl(startURL, pc)
	}()

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			zeno.IndexTF(pc)
		}()
	}
	wg.Wait()
	zeno.IndexDF()
}
