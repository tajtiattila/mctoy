package main

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"
)

func TestConfigCrypto(t *testing.T) {
	rand.Seed(time.Now().Unix())
	c := config{make(map[string]string)}
	for i := 0; i < 10; i++ {
		b := make([]byte, rand.Int()&0xff)
		for i := range b {
			b[i] = byte(rand.Int31())
		}
		plaintext0 := string(b)
		ciphertext := c.Encrypt(plaintext0)
		plaintext1 := c.Decrypt(ciphertext)
		if plaintext0 != plaintext1 {
			t.Error("crypto error:", hex.EncodeToString(b))
		} else {
			t.Log("crypto OK:", hex.EncodeToString(b))
		}
	}
}
