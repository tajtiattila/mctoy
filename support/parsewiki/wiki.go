package main

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

type TableCell struct {
	r, c, rowspan, colspan int
	text                   string
}

func (tc *TableCell) String() string { return tc.text }

type Table struct {
	nr, nc int
	data   []*TableCell
	nh, nf int
}

func (t *Table) NRows() int    { return t.nr }
func (t *Table) NCols() int    { return t.nc }
func (t *Table) NHeaders() int { return t.nh }
func (t *Table) NFooters() int { return t.nf }
func (t *Table) Cell(r, c int) *TableCell {
	if 0 <= r && r < t.nr && 0 <= c && c < t.nc {
		return t.data[r*t.nc+c]
	}
	return nil
}
func (t *Table) SetCell(r, c int, d *TableCell) {
	t.extend(r+1, c+1)
	t.data[r*t.nc+c] = d
}
func (t *Table) Clone() *Table {
	return t.SubTable(0, 0, t.NRows(), t.NCols())
}
func (t *Table) Data() *Table {
	if t.nh == 0 && t.nf == 0 {
		return t
	}
	return t.SubTable(t.NHeaders(), 0, t.NRows()-(t.NHeaders()+t.NFooters()), t.NCols())
}

func (t *Table) SubTable(or, oc, nr, nc int) *Table {
	tc := new(Table)
	tc.resize(nr, nc)
	for y := 0; y < nr; y++ {
		for x := 0; x < nc; x++ {
			tc.set(y, x, t.Cell(or+y, oc+x))
		}
	}
	return tc
}

func (t *Table) set(r, c int, d *TableCell) {
	t.data[r*t.nc+c] = d
}
func (t *Table) resize(r, c int) {
	if c == t.nc || (r == t.nr && r == 1) {
		t.nr, t.nc = r, c
		if t.nr*t.nc > len(t.data) {
			t.data = append(t.data, make([]*TableCell, t.nr*t.nc)...)
		}
		t.data = t.data[:t.nr*t.nc]
		return
	}
	t0 := Table{t.nr, t.nc, t.data, 0, 0}
	t.nr, t.nc = r, c
	t.data = make([]*TableCell, t.nr*t.nc)
	for y := 0; y < t.nr; y++ {
		for x := 0; x < t.nc; x++ {
			t.set(y, x, t0.Cell(y, x))
		}
	}
}
func (t *Table) extend(r, c int) {
	xr, xc := t.nr, t.nc
	if r > xr {
		xr = r
	}
	if c > xc {
		xc = c
	}
	if t.nc != xc || t.nr != xr {
		t.resize(xr, xc)
	}
}

////////////////////////////////////////////////////////////////////////////////

type cellFmt struct {
	rowspan, colspan int
}

func getInt(vs string, minval int) int {
	vs = strings.TrimSpace(vs)
	if len(vs) > 2 && vs[0] == '"' && vs[len(vs)-1] == '"' {
		vs = vs[1 : len(vs)-1]
	}
	iv, err := strconv.ParseInt(vs, 10, 32)
	i := minval
	if err == nil && int(iv) > minval {
		i = int(iv)
	}
	return i
}
func parseFmt(fs string) cellFmt {
	f := cellFmt{1, 1}
	for _, s := range strings.Fields(fs) {
		if p := strings.IndexRune(s, '='); p != -1 {
			k, v := s[:p], s[p+1:]
			switch k {
			case "rowspan":
				f.rowspan = getInt(v, 1)
			case "colspan":
				f.colspan = getInt(v, 1)
			}
		}
	}
	return f
}

type TableBuilder struct {
	row, col int
	t        *Table
}

func (tb *TableBuilder) tableCell(ch byte, fs, text string) {
	if tb.t == nil {
		tb.t = new(Table)
	}
	if tb.col == 0 && ch == '!' {
		if tb.row <= tb.t.nh {
			tb.t.nh++
		} else {
			tb.t.nf++
		}
	}
	f := parseFmt(fs)
	tc := &TableCell{tb.col, tb.row, f.rowspan, f.colspan, text}
	for dr := 0; dr < f.rowspan; dr++ {
		for dc := 0; dc < f.colspan; dc++ {
			tb.t.SetCell(tb.row+dr, tb.col+dc, tc)
		}
	}
	tb.col += f.colspan
	tb.skip()
}

func (tb *TableBuilder) nextrow(l string) {
	if tb.t == nil {
		return
	}
	if tb.col == 0 {
		hascell := false
		for x := 0; x < tb.t.NCols(); x++ {
			if hascell = tb.t.Cell(x, tb.row) != nil; hascell {
				break
			}
		}
		if !hascell {
			return
		}
	}
	tb.row++
	tb.col = 0
	tb.skip()
}

func (tb *TableBuilder) skip() {
	for tb.t.Cell(tb.row, tb.col) != nil {
		tb.col++
	}
}

func (tb *TableBuilder) caption(s string) {
}

////////////////////////////////////////////////////////////////////////////////

/*
func WikiTitle string

func state func(*ctx) state

func startLine(c *ctx) state {
	switch {
	case t.has("{|"):
		t.tableBuilder = new(TableBuilder)
		t.eatLine()
		return startTableLine
	case t.has("=="):
		title = strings.TrimSpace(strings.Trim(strings.TrimSpace(t.eatLine()), "="))
		t.emit(WikiTitle(title))
		return startLine
	}
	return normalText
}

func startTableLine(c *ctx) state {
	switch {
	case strings.HasPrefix(l, "|+"):
		l := t.eatLine()
		t.tableBuilder.caption(l[2:])
	case strings.HasPrefix(l, "|-"):
		l := t.eatLine()
		t.tableBuilder.nextrow(l[2:])
	case len(l) != 0 && (l[0] == '|' || l[0] == '!'):
		ch := l[0]
		sep := string([]byte{ch, ch})
		l = l[1:]
		for _, c := range strings.Split(l, sep) {
			p := strings.IndexRune(c, '|')
			var cf, ct string
			if p >= 0 {
				cf, ct = c[:p], c[p+1:]
			} else {
				cf, ct = "", c
			}
			tb.tableCell(ch, cf, strings.TrimSpace(ct))
		}
	}
}

func inlineText(c *ctx) state {
}

*/

////////////////////////////////////////////////////////////////////////////////

func parseWikiTable(tab []string) (*Table, error) {
	if len(tab) < 2 || !strings.HasPrefix(tab[0], "{|") || tab[len(tab)-1] != "|}" {
		return nil, errors.New("MalformedTable")
	}
	tab = tab[1 : len(tab)-1]
	var tb TableBuilder
	for _, l := range tab {
		switch {
		case strings.HasPrefix(l, "|+"):
			tb.caption(l[2:])
		case strings.HasPrefix(l, "|-"):
			tb.nextrow(l[2:])
		case len(l) != 0 && (l[0] == '|' || l[0] == '!'):
			ch := l[0]
			sep := string([]byte{ch, ch})
			l = l[1:]
			for _, c := range strings.Split(l, sep) {
				p := strings.IndexRune(c, '|')
				l := strings.Index(c, "[[")
				if l >= 0 && p > l {
					p = -1
				}
				var cf, ct string
				if p >= 0 {
					cf, ct = c[:p], c[p+1:]
				} else {
					cf, ct = "", c
				}
				tb.tableCell(ch, cf, strings.TrimSpace(ct))
			}
		}
	}
	return tb.t, nil
}

type wikiProc interface {
	Title(string) error
	Table(wtable *Table) error
}

func procWikiTables(r io.Reader, p wikiProc) error {
	scanner := bufio.NewScanner(r)
	tab := make([]string, 0, 20)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "=="):
			title := strings.TrimSpace(strings.Trim(strings.TrimSpace(line), "="))
			if err := p.Title(title); err != nil {
				return err
			}
		case strings.HasPrefix(line, "{|"):
			if len(tab) != 0 {
				return errors.New("TableInsideTable")
			}
			tab = append(tab, line)
		case strings.HasPrefix(line, "|}"):
			tab = append(tab, line)
			wt, err := parseWikiTable(tab)
			if err == nil {
				p.Table(wt)
			}
			if err != nil {
				return err
			}
			tab = tab[:0]
		default:
			if len(tab) != 0 {
				tab = append(tab, line)
			}
		}
	}
	return scanner.Err()
}
