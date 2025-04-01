package main

import (
	"github.com/josuetorr/nomad/internal/db"
)

const startURL = "https://wikipedia.org/wiki/meme"

func main() {
	kv := db.NewKV("/tmp/badger/nomad")
}
