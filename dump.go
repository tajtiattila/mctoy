package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const PACKETDEBUG = true

type Dumper struct {
	l      io.Writer
	b1, b2 bytes.Buffer
}

func MakeDumper(l io.Writer) Dumper {
	return Dumper{l: l}
}
func (d Dumper) line(pfx string, err error) {
	s := ""
	if err != nil {
		s = " ERROR: " + err.Error()
	}
	fmt.Fprint(d.l, pfx, s, "\n")
}

func (d Dumper) bytes(buf []byte) {
	sfx := ""
	trunc := 256
	if len(buf) > 2*trunc {
		sfx = fmt.Sprint("... ", len(buf)-trunc, " more bytes")
		buf = buf[:trunc]
	}
	lmod := 16
	e := func(force bool) {
		if d.b2.Len() != 0 {
			if !force && d.b2.Len() < lmod {
				return
			}
			for i := d.b2.Len(); i < lmod; i++ {
				fmt.Fprint(&d.b1, "   ")
			}
			fmt.Fprint(d.l, "    ")
			io.Copy(d.l, &d.b1)
			fmt.Fprint(d.l, "  ")
			io.Copy(d.l, &d.b2)
			fmt.Fprint(d.l, "\n")
			d.b1.Reset()
			d.b2.Reset()
		}
	}
	for i := 0; i < len(buf); i++ {
		fmt.Fprintf(&d.b1, "%02x ", buf[i])
		r := rune(buf[i])
		if r < 32 || 127 <= r {
			r = '.'
		}
		fmt.Fprintf(&d.b2, "%c", r)
		if (i+1)%lmod == 0 {
			e(false)
		}
	}
	e(true)
	if sfx != "" {
		fmt.Println(sfx)
	}
}

type DebugReader struct {
	Dumper
	r io.Reader
}

func NewDebugReader(r io.Reader, l io.Writer) *DebugReader {
	return &DebugReader{MakeDumper(l), r}
}
func (d *DebugReader) Read(buf []byte) (n int, err error) {
	n, err = d.r.Read(buf)
	d.line(fmt.Sprintf("<- %d/%d", n, len(buf)), err)
	d.bytes(buf[:n])
	return
}

type DebugWriter struct {
	Dumper
	w io.Writer
}

func NewDebugWriter(w io.Writer, l io.Writer) *DebugWriter {
	return &DebugWriter{MakeDumper(l), w}
}
func (d *DebugWriter) Write(buf []byte) (n int, err error) {
	n, err = d.w.Write(buf)
	d.line(fmt.Sprintf("-> %d/%d", n, len(buf)), err)
	d.bytes(buf)
	return
}

func dumpJson(i interface{}) string {
	d, err := json.Marshal(i)
	if err != nil {
		return "JsonMarshalError: " + err.Error()
	}
	buf := bytes.NewBuffer(nil)
	err = json.Indent(buf, d, "  ", "  ")
	if err != nil {
		return string(d)
	}
	return string(buf.Bytes())
}
