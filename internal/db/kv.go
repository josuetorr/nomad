package db

import (
	"log"

	badger "github.com/dgraph-io/badger/v4"
)

type KV struct {
	db *badger.DB
}

func NewKV(path string) *KV {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		log.Fatalf("Failed to open badger: %s", err)
	}
	return &KV{
		db: db,
	}
}

func (kv *KV) Put(key string, value []byte) error {
	return kv.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value)
		if err := txn.SetEntry(e); err != nil {
			return err
		}
		return nil
	})
}

func (kv *KV) Get(key string) ([]byte, error) {
	var val []byte
	kv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		item.ValueCopy(val)
		return nil
	})
	return val, nil
}

func (kv *KV) Exists(key string) bool {
	val, err := kv.Get(key)
	if err != nil {
		return false
	}

	return len(val) > 0
}

func (kv *KV) Delete(key string) error {
	panic("delete not implemented")
}

func (kv *KV) Close() error {
	return kv.db.Close()
}
