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
