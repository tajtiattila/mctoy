package main

import (
	"io"
)

// roundBuf keeps the last part of the data
// written to it. First Write into it, then Read
// to access the last bits of your data. When reading,
// the first cached line or fraction of a line is
// dropped.
type roundBuf struct {
	buf  []byte
	r, w int
	full bool
	read bool
}

func NewRoundBuf() io.ReadWriter {
	return &roundBuf{buf: make([]byte, 65536)}
}

func (rb *roundBuf) Write(b []byte) (n int, err error) {
	if rb.read {
		return 0, io.EOF
	}
	n = len(b)
	if len(rb.buf) <= n {
		copy(rb.buf, b)
		rb.w = 0
		return
	}
	nc := copy(rb.buf[rb.w:], b)
	b = b[nc:]
	rb.w += nc
	if rb.w == len(rb.buf) {
		rb.w, rb.full = 0, true
	}
	rb.w = copy(rb.buf, b)
	return
}

func (rb *roundBuf) Read(b []byte) (n int, err error) {
	if !rb.read {
		rb.read = true
		if !rb.full {
			rb.r = 0
		}
		for {
			ch := rb.buf[rb.r]
			rb.r++
			if rb.r == len(rb.buf) {
				rb.r = 0
			}
			if ch == '\n' || rb.r == rb.w {
				break
			}
		}
	}
	if rb.w < rb.r {
		n = copy(b, rb.buf[rb.r:])
		rb.r, b = rb.r+n, b[n:]
		if rb.r == len(rb.buf) {
			rb.r = 0
		}
	}
	if rb.r == rb.w {
		if n == 0 {
			err = io.EOF
		}
		return
	}
	nc := copy(b, rb.buf[:rb.w])
	n, rb.r, b = n+nc, rb.r+nc, b[nc:]
	return
}
