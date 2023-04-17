package main

import (
	"crypto/sha1"
	"encoding/hex"
)

func EncryptedPassword(password string) string {
	h := sha1.New()
	h.Write([]byte(password))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}
