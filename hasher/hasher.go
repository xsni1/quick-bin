package hasher

import (
	"math/rand"
	"strings"
)

func Hasher(length int) string {
	const chars = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"
	var sb strings.Builder

	for i := 0; i < length; i++ {
		sb.WriteByte(chars[rand.Intn(len(chars))])
	}

	hash := sb.String()
	return hash
}
