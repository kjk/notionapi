package notionapi

import (
	"testing"

	"github.com/kjk/common/assert"
)

func TestExtractNoDashIDFromNotionURL(t *testing.T) {
	tests := [][]string{
		{
			"https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6",
			"ea07db1b9bff415ab180b0525f3898f6",
		},
		{
			"https://www.notion.so/kjkpublic/Type-assertion-e945ebc2e0074ce49cef592e6c0f956e",
			"e945ebc2e0074ce49cef592e6c0f956e",
		},
		{
			"https://notion.so/f400553890d34185ba795870807c2616",
			"f400553890d34185ba795870807c2616",
		},
		{
			"f400553890d34185ba795870807c2617",
			"f400553890d34185ba795870807c2617",
		},
		{
			"f400553890d34185ba795870-807c2618",
			"f400553890d34185ba795870807c2618",
		},
		{
			"https://www.notion.so/kjkpublic/Empty-interface-c3315892508248fdb19b663bf8bff028#0500145a75da4464bca0d25da19af112",
			"c3315892508248fdb19b663bf8bff028",
		},
	}
	for _, tc := range tests {
		got := ExtractNoDashIDFromNotionURL(tc[0])
		exp := tc[1]
		assert.Equal(t, exp, got)
	}
}
