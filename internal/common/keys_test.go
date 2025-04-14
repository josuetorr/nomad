package common_test

import (
	"fmt"
	"testing"

	"github.com/josuetorr/nomad/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestTermKey_Successful(t *testing.T) {
	expected := fmt.Sprintf("term%stest", common.KeySep)
	result := common.TermKey("test")

	assert.Equal(t, expected, result)
}

func TestDocKey_Successful(t *testing.T) {
	expected := fmt.Sprintf("doc%swikipedia.com", common.KeySep)
	result := common.DocKey("wikipedia.com")

	assert.Equal(t, expected, result)
}

func TestKeyParts_Successful(t *testing.T) {
	expected := []string{"term", "test", "hello"}
	result := common.KeyParts(common.TermKey("test", "hello"))

	assert.Equal(t, expected, result)
}
