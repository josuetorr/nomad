package common

type Storer interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	IteratePrefix(key string, fn func(val []byte) error) error
	Exists(key string) bool
	Delete(key string) error
	Close() error
}
