package db

type KV struct{}

func NewKV() *KV {
	return &KV{}
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
