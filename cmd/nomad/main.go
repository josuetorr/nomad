package main

import (
	"fmt"
	"sync"

	"github.com/josuetorr/nomad/internal/pipeline"
)

const startURL = "https://wikipedia.org/wiki/meme"

func main() {
	pl := pipeline.Init()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pl.Crawl(startURL)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for dj := range pl.DocJobs {
			for _, d := range dj {
				fmt.Printf("%s\n", d.Url)
			}
		}
	}()
	wg.Wait()
	println("done")
}
