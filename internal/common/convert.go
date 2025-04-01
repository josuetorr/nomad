package common

import (
	"bytes"
	"encoding/binary"
)

func Uint64ToBytes(val uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, val)
	return b
}

func BytesToUint64(b []byte) (uint64, error) {
	var num uint64
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.LittleEndian, &num)
	if err != nil {
		return 0, err
	}
	return num, nil
}
