package security

import (
	"crypto/rand"
	"github.com/satori/go.uuid"
)

func UUID() string {
	return uuid.NewV4().String()
}

func RandString(length int) string {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if length == 0 {
		return ""
	}
	clen := len(letters)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("Error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue // Skip this number to avoid modulo bias.
			}
			b[i] = letters[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
