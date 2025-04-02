package db

type KVStorer interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	BatchWrite(fn func(w KVWriter)) error
	IteratePrefix(key string, fn func(key []byte, val []byte) error) error
	Exists(key string) bool
	Delete(key string) error
	Close() error
}

type KVWriter interface {
	Set(key []byte, val []byte) error
}
