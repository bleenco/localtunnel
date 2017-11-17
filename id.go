package localtunnel

import (
	"crypto/rand"
	"fmt"
)

func randID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
