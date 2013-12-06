package main

import (
	"fmt"
	"os"
	"strings"
)

type McPacket struct {
	ID     byte
	Fields []McField
}

type McField struct {
	Name  string
	Type  string
	Notes string
}

func cellString(t *Table, r, c int) string {
	if tc := t.Cell(r, c); tc != nil {
		return tc.String()
	}
	return ""
}

func dumpTable(title string, t *Table) error {
	fmt.Printf("%s:\n", title)
	for y := 0; y < t.NRows(); y++ {
		sep := " | "
		if y < t.NHeaders() || y >= t.NRows()-t.NFooters() {
			sep = " ! "
		}
		for x := 0; x < t.NCols(); x++ {
			fmt.Print(sep, t.Cell(y, x))
		}
		fmt.Println()
	}
	fmt.Println()
	return nil
}

func arrayType(n, j string) string {
	j = strings.TrimSpace(j)
	if len(j) > 0 && j[len(j)-1] == 's' {
		j = j[:len(j)-1]
	}
	return "[]" + mapType(n, j)
}

func mapType(n, j string) string {
	lj := strings.ToLower(j)
	switch lj {
	case "varint":
		return "uint"
	case "int":
		return "int32"
	case "unsigned int":
		return "int32"
	case "short":
		return "int16"
	case "unsigned short":
		return "int16"
	case "byte":
		return "int8"
	case "unsigned byte":
		return "uint8"
	case "long":
		return "int64"
	case "unsigned long":
		return "uint64"
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "string", "bool":
		return lj
	}
	if strings.HasPrefix(lj, "array of") {
		return arrayType(n, j[8:])
	} else if strings.HasSuffix(lj, "array") {
		return arrayType(n, j[:len(j)-5])
	}
	return makeName(j)
}

func makeName(base string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '-', ' ':
			return -1
		}
		return r
	}, strings.Title(base))
}

type procMcProto struct {
	title string
	elems []interface{}
}

func (m *procMcProto) Title(t string) error {
	m.title = t
	if t == "Serverbound" || t == "Clientbound" {
		m.elems = append(m.elems, "//----- ", t, "\n")
	}
	return nil
}
func (m *procMcProto) Table(t *Table) error {
	td := t.Data()
	if t.NHeaders() != 0 && cellString(t, 0, 0) == "Packet ID" {
		fmt.Println("//", td.Cell(0, 0), "=", m.title)
		hexid := td.Cell(0, 0)
		name := makeName(m.title)
		m.elems = append(m.elems, fmt.Sprint("// ", name, "\n"))
		fmt.Printf("type %s struct {\n", name)
		for r := 0; r < t.NRows(); r++ {
			fname := makeName(cellString(td, r, 1))
			ftype := mapType(fname, cellString(td, r, 2))
			fcomm := cellString(td, r, 3)
			if fcomm != "" {
				fcomm = " // " + fcomm
			}
			l := strings.TrimSpace(fmt.Sprintf("%-20s  %-10s%s", fname, ftype, fcomm))
			switch l {
			case "//", "":
			default:
				fmt.Print("  ", l, "\n")
			}
		}
		fmt.Printf("}\nfunc (%s) Id() uint { return %s }\n\n", name, hexid)
	}
	return nil
}

func main() {
	f, err := os.Open("proto.wiki")
	check(err)
	defer f.Close()
	//err = procWikiTables(f, dumpTable)
	var p procMcProto
	err = procWikiTables(f, &p)
	check(err)
	fmt.Print(p.elems...)
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
