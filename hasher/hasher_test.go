package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher(t *testing.T) {
	const chars = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"
	length := 5
	result := Hasher(length)

	assert.Len(t, result, length)

	for _, c := range result {
		assert.Contains(t, chars, string(c))
	}
}
