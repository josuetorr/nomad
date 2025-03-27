package common

import (
	"bytes"
	"compress/gzip"
)

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	defer w.Close()
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
