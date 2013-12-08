package net

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"math/rand"
	"testing"
)

func getRandomBytes(lo, up int) (b []byte) {
	if lo != up {
		b = make([]byte, rand.Int()%(up-lo)+lo)
	} else {
		b = make([]byte, lo)
	}
	for i := range b {
		b[i] = byte(rand.Int31())
	}
	return
}

type encdecpair struct {
	e cipher.Stream
	d cipher.Stream
}

func tc(t *testing.T, p encdecpair, plaintext []byte) []byte {
	coded := new(bytes.Buffer)
	w := cipher.StreamWriter{
		W: coded,
		S: p.e,
	}
	r := cipher.StreamReader{
		R: coded,
		S: p.d,
	}
	n, err := w.Write(plaintext)
	if err != nil {
		t.Error("Encode error")
		return nil
	}
	if n != len(plaintext) {
		t.Error("nwritten != len(plaintext)")
		return nil
	}
	ciphertext := make([]byte, coded.Len())
	copy(ciphertext, coded.Bytes())
	decoded := make([]byte, len(plaintext))
	n, err = io.ReadFull(r, decoded)
	if err != nil {
		t.Error("Decode error")
		return nil
	}
	if n != len(decoded) {
		t.Error("nread != len(plaintext)")
		return nil
	}
	if coded.Len() != 0 {
		t.Error("bytes remaining in coded buffer")
	}
	return ciphertext
}

func mkCFB8(c cipher.Block, iv []byte) encdecpair {
	return encdecpair{
		NewCFB8Encrypter(c, iv),
		NewCFB8Decrypter(c, iv),
	}
}

func mkCFB(c cipher.Block, iv []byte) encdecpair {
	return encdecpair{
		cipher.NewCFBEncrypter(c, iv),
		cipher.NewCFBDecrypter(c, iv),
	}
}

func TestCFBCompare(t *testing.T) {
	for i := 0; i < 10; i++ {
		secret := getRandomBytes(16, 16)
		ciph, err := aes.NewCipher(secret)
		if err != nil {
			t.Fatal("Can't create cipher")
		}
		plaintext := getRandomBytes(3, 128)
		c1 := tc(t, mkCFB8(ciph, secret), plaintext)
		c2 := tc(t, mkCFB(ciph, secret), plaintext)
		if !bytes.Equal(c1, c2) {
			t.Logf("go != mc\n  %x\n  %x\n", c1, c2)
		}
	}
}
