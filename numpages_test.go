package pdfrender

import (
	"os"
	"testing"
)

// NumPages is what \includepdf resolves a page RANGE against: pages={2-5} and
// pages=- name a count only the document knows. Counting "/Type /Page" in the bytes
// reads 0 on a PDF whose page tree is in an object stream — which is what a modern
// writer produces — and a range bounded against 0 silently includes nothing.
func TestNumPages(t *testing.T) {
	for _, c := range []struct {
		file string
		want int
	}{
		{"testdata/three-pages.pdf", 3},
	} {
		data, err := os.ReadFile(c.file)
		if err != nil {
			t.Skipf("%s: %v", c.file, err)
		}
		if got := NumPages(data); got != c.want {
			t.Errorf("NumPages(%s) = %d, want %d", c.file, got, c.want)
		}
	}
}

// A file that is not a PDF reports 0 rather than panicking: the caller treats an
// unknown count as "do not clamp", which is the safe direction.
func TestNumPagesOnRubbish(t *testing.T) {
	if got := NumPages([]byte("not a pdf at all")); got != 0 {
		t.Errorf("= %d, want 0", got)
	}
}
