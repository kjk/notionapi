package notionapi

import (
	"bytes"
	"fmt"
	"io"
)

type writer struct {
	level int
	w     io.Writer
}

func (w *writer) writeString(s string) {
	_, _ = io.WriteString(w.w, s)
}

func (w *writer) writeLevel() {
	for n := 0; n < w.level; n++ {
		w.writeString("  ")
	}
}

func (w *writer) block(block *Block) {
	if block == nil {
		return
	}
	w.writeLevel()
	s := fmt.Sprintf("%s %s alive=%v\n", block.Type, block.ID, block.Alive)
	w.writeString(s)
	w.level++
	for _, child := range block.Content {
		w.block(child)
	}
	w.level--
}

// Dump writes a simple representation of Page to w. A debugging helper.
func Dump(w io.Writer, page *Page) {
	wr := writer{w: w}
	wr.block(page.Root())
}

// DumpToString returns a simple representation of Page as a string.
// A debugging helper.
func DumpToString(page *Page) string {
	buf := &bytes.Buffer{}
	Dump(buf, page)
	return buf.String()
}
