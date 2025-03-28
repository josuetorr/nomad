package common_test

import (
	"testing"

	"github.com/josuetorr/nomad/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestCompression_Successful(t *testing.T) {
	expected := "hello, world!"
	c, _ := common.Compress([]byte(expected))
	result, _ := common.Decompress(c)

	assert.Equal(t, expected, string(result))
}
