package main

import (
	"fmt"

	"github.com/josuetorr/nomad/internal/db"
	"github.com/josuetorr/nomad/internal/spidey"
)

func main() {
	const entryPoint = "https://wikipedia.org/wiki/meme"
	kv := db.NewKV("/tmp/badger")
	crawlCh := make(chan spidey.CrawledPage, 100)
	spider := spidey.NewSpidey(kv, crawlCh)
	go spider.Crawl(entryPoint)

	data := <-crawlCh
	fmt.Printf("crawled data: %+v\n", data.Indexable)
}
