package notionapi

import "testing"

func TestNormalizeID(t *testing.T) {
	var tests = []struct {
		s   string
		exp string
	}{
		{"2131b10cebf64938a1277089ff02dbe4", "2131b10c-ebf6-4938-a127-7089ff02dbe4"},
		{"2131b10c-ebf6-4938-a127-7089ff02dbe4", "2131b10c-ebf6-4938-a127-7089ff02dbe4"},
		{"2131b", "2131b"},
	}
	for _, test := range tests {
		got := ToDashID(test.s)
		if got != test.exp {
			t.Errorf("s: %s got: %s, expected: %s\n", test.s, got, test.exp)
		}
	}
}
