package db

import (
	"fmt"
	"log"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

type KV struct {
	db *badger.DB
}

type MergeFunc = badger.MergeFunc

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

func (kv *KV) ReadWrite(k string, mapper func([]byte) ([]byte, error)) error {
	kv.db.Update(func(txn *badger.Txn) error {
		k := []byte(k)
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		v, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		mappedv, err := mapper(v)
		if err != nil {
			return err
		}

		return txn.Set(k, mappedv)
	})
	return nil
}

func (kv *KV) BatchWrite(fn func(w KVWriter) error) error {
	wb := kv.db.NewWriteBatch()
	defer wb.Cancel()
	if err := fn(wb); err != nil {
		return err
	}
	return wb.Flush()
}

func (kv *KV) Merge(key []byte, op MergeFunc, dur time.Duration, vals ...[]byte) ([]byte, error) {
	mo := kv.db.GetMergeOperator(key, op, dur)
	defer mo.Stop()
	for _, v := range vals {
		mo.Add(v)
	}

	return mo.Get()
}

func (kv *KV) IteratePrefix(prefix string, fn func(key []byte, val []byte) error) error {
	return kv.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		p := []byte(prefix)
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				return fn(key, val)
			})
			if err != nil {
				fmt.Printf("Failed to get value for: %s. err: %s\n", string(key), err)
			}
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
