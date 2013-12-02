// thank you, Toqueteos!
// https://gist.github.com/toqueteos/5372776
package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"io"
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

func GenerateSharedSecret() ([]byte, error) {
	sharedSecret := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, sharedSecret); err != nil {
		return nil, err
	}
	return sharedSecret, nil
}

func GenerateV4UUID() (string, error) {
	src := bufio.NewReaderSize(rand.Reader, 1)
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
