package common

type Storer interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	Exists(key string) bool
	Delete(key string) error
	Close() error
}
