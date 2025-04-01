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

func (kv *KV) Put(k string, v []byte) error {
	return kv.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(k), v)
		return txn.SetEntry(e)
	})
}

func (kv *KV) Get(k string) ([]byte, error) {
	var v []byte
	if err := kv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		v, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return v, nil
}

func (kv *KV) BatchWrite() {
}

func (kv *KV) IteratePrefix(prefix string, fn func(val []byte) error) error {
	return kv.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		p := []byte(prefix)
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			item := it.Item()
			return item.Value(fn)
		}
		return nil
	})
}

func (kv *KV) Exists(k string) bool {
	val, err := kv.Get(k)
	if err != nil {
		return false
	}

	return len(val) > 0
}

func (kv *KV) Delete(k string) error {
	panic("delete not implemented")
}

func (kv *KV) Close() error {
	return kv.db.Close()
}
