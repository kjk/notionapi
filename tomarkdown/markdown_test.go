package tomarkdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownFileNameForPage(t *testing.T) {
	tests := [][]string{
		{"Blendle's Employee Handbook", "3b617da409454a52bc3a920ba8832bf7", "Blendle-s-Employee-Handbook-3b617da4-0945-4a52-bc3a-920ba8832bf7.md"},
		{"To Do/Read in your first week", "5fea9664-0720-4d90-80a5-b989360b205f", "To-Do-Read-in-your-first-week-5fea9664-0720-4d90-80a5-b989360b205f.md"},
		{"DNA-&-culture", "9cc14382-e3c3-4037-bf80-a4936a9b6674", "DNA-culture-9cc14382-e3c3-4037-bf80-a4936a9b6674.md"},
		{"General & practical ", "6d25f4e5-3b91-4df6-8630-c98ea5523692", "General-practical-6d25f4e5-3b91-4df6-8630-c98ea5523692.md"},
		{"Time off: holidays and national holidays", "d0464f97-6364-48fd-8dab-5497f68394c2", "Time-off-holidays-and-national-holidays-d0464f97-6364-48fd-8dab-5497f68394c2.md"},
		{"#letstalkaboutstress", "94a2bcc4-7fde-4dab-9229-68733b9a2a94", "letstalkaboutstress-94a2bcc4-7fde-4dab-9229-68733b9a2a94.md"},
		{"The Matrix™ (job profiles)", "f495439c-3d54-409c-a714-fc3c7cc5711f", "The-Matrix-job-profiles-f495439c-3d54-409c-a714-fc3c7cc5711f.md"},
	}
	for _, test := range tests {
		got := markdownFileName(test[0], test[1])
		assert.Equal(t, test[2], got)
	}
}
