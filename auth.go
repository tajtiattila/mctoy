// thank you, Toqueteos!
// https://gist.github.com/toqueteos/5372776
package main

import (
	"bufio"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand"
	"strings"
)

func McDigest(hash []byte) string {
	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(hex.EncodeToString(hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

// little endian
func twosComplement(p []byte) {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// return a ctypto-inited pseudo random generator
func NewRandom() (*rand.Rand, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(crand.Reader, buf); err != nil {
		return nil, err
	}
	seed := int64(binary.BigEndian.Uint64(buf))
	return rand.New(rand.NewSource(seed)), nil
}
func MustNewRandom() *rand.Rand {
	r, e := NewRandom()
	if e != nil {
		panic(e)
	}
	return r
}

type randReader struct {
	r   *rand.Rand
	buf []byte
	pos int
}

func (r *randReader) Read(b []byte) (nread int, err error) {
	nread = len(b)
	for len(b) != 0 {
		if r.pos == 0 {
			u1, u2 := uint64(r.r.Int63()), uint64(r.r.Int63())
			u := u1 ^ (u2 << 1)
			binary.BigEndian.PutUint64(r.buf, u)
		}
		n := copy(b, r.buf[r.pos:])
		r.pos, b = (r.pos+n)&7, b[n:]
	}
	return
}
func NewRandReader(r *rand.Rand) io.Reader {
	return &randReader{r, make([]byte, 8), 0}
}

var (
	authRand   = MustNewRandom()
	authReader = NewRandReader(authRand)
)

func GenerateSharedSecret() ([]byte, error) {
	sharedSecret := make([]byte, 16)
	if _, err := io.ReadFull(authReader, sharedSecret); err != nil {
		return nil, err
	}
	return sharedSecret, nil
}

func GenerateV4UUID() (string, error) {
	src := bufio.NewReaderSize(authReader, 1)
	b := []byte("xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx")
	vhex := "0123456789abcdef"
	for i := range b {
		switch b[i] {
		case 'x', 'y':
			byt, err := src.ReadByte()
			if err != nil {
				return "", err
			}
			digit := int(byt) % 16
			if b[i] == 'y' {
				digit = 8 + digit%4
			}
			b[i] = vhex[digit]
		}
	}
	return string(b), nil
}
