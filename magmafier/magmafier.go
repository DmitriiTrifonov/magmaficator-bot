package magmafier

import (
	"math/rand"
	"time"
)

func MakeKeyFromString(s string) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 32)
	if s == "" {
		_, _ = rand.Read(b)
	} else {
		b = []byte(s)
	}
	if len(b) < 32 {
		rest := make([]byte, 32-len(b))
		b = append(b, rest...)
	}
	if len(b) > 32 {
		b = b[:32]
	}
	return b
}
