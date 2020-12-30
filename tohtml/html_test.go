package tohtml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTMLFileNameForPage(t *testing.T) {
	tests := [][]string{
		{"Blendle's Employee Handbook", "Blendle s Employee Handbook.html"},
	}
	for _, test := range tests {
		got := htmlFileName(test[0])
		assert.Equal(t, test[1], got)
	}
}

func TestFmtNumberWithCommas(t *testing.T) {
	tests := []string{
		"1345.48", "1,345.48",
		"", "",
		"0", "0",
		".32", ".32",
		"345", "345",
		"3.12", "3.12",
		"3467893.2213", "3,467,893.2213",
	}
	n := len(tests) / 2
	for i := 0; i < n; i++ {
		s := tests[i*2]
		got := fmtNumberWithCommas(s)
		exp := tests[i*2+1]
		assert.Equal(t, exp, got)
	}
}
