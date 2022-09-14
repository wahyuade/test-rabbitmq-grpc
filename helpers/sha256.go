package helpers

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	memory      = 1
	iterations  = 1
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

func Hash(raw string) (string, error) {
	h := sha256.New()
	h.Write([]byte(raw))
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs), nil

}

func MatchHash(raw string, encodedHash string) (bool, error) {
	hash, err := Hash(raw)
	return hash == encodedHash, err

}
