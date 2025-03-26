package db

import (
	"log"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/josuetorr/nomad/internal/common"
)

type KV struct {
	db *badger.DB
}

func NewKV(cfg common.Config) *KV {
	db, err := badger.Open(badger.DefaultOptions(cfg.BadgerDir))
	if err != nil {
		log.Fatalf("Failed to open badger: %s", err)
	}
	return &KV{
		db: db,
	}
}

func (kv *KV) Put(key, value []byte) error {
	panic("put not implemented")
}

func (kv *KV) Get(key []byte) error {
	panic("get not implemented")
}

func (kv *KV) Delete(key []byte) error {
	panic("delete not implemented")
}

func (kv *KV) Close() error {
	panic("close not implemented")
}
