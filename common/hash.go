package common

import (
	"crypto/sha256"
)

const (
	HashSize = 32
)

type Hash [HashSize]byte

func DoubleHash(b []byte) Hash {
	first := sha256.Sum256(b)
	return Hash(sha256.Sum256(first[:]))
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashSize:]
	}
	copy(h[HashSize-len(b):], b)
}
