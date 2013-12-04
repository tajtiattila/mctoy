/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"crypto/cipher"
)

func NewCFB8Decrypter(c cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(c, iv, true)
}

func NewCFB8Encrypter(c cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(c, iv, false)
}

type cfb8 struct {
	c            cipher.Block
	blockSize    int
	buf, iv, out []byte
	de           bool
}

const CircularBufferLen = 256

func newCFB8(c cipher.Block, iv []byte, decrypt bool) *cfb8 {
	if len(iv) != c.BlockSize() {
		panic("newCFB8: IV length must equal block size")
	}
	bsiz, n := c.BlockSize(), CircularBufferLen
	if n < bsiz*4 {
		n = bsiz * 4
	}
	buf := make([]byte, n+bsiz)
	copy(buf, iv)
	return &cfb8{
		c:         c,
		blockSize: c.BlockSize(),
		buf:       buf,
		iv:        buf[:bsiz],
		out:       make([]byte, bsiz),
		de:        decrypt,
	}
}

func (x *cfb8) XORKeyStream(dst, src []byte) {
	bsiz := x.c.BlockSize()
	for i := 0; i < len(src); i++ {
		b := src[i]
		x.c.Encrypt(x.out, x.iv)
		b = b ^ x.out[0]

		if cap(x.iv) > bsiz {
			x.iv = x.iv[1 : bsiz+1]
		} else {
			copy(x.buf, x.iv[1:])
			x.iv = x.buf[:bsiz]
		}

		if x.de {
			x.iv[bsiz-1] = src[i]
		} else {
			x.iv[bsiz-1] = b
		}
		dst[i] = b
	}
}
