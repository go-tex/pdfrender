// Copyright (c) the go-tex/pdfrender authors.
// SPDX-License-Identifier: BSD-3-Clause

package pdfrender

import (
	"os"
	"testing"
)

// nonWhiteFraction is the share of pixels that are not near-white — a proxy for
// "the page actually drew something".
func nonWhiteFraction(t *testing.T, pdf []byte) float64 {
	t.Helper()
	img, err := Rasterize(pdf, 150)
	if err != nil {
		t.Fatalf("Rasterize: %v", err)
	}
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		t.Fatal("empty image")
	}
	var n int
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r < 0xf000 || g < 0xf000 || bl < 0xf000 {
				n++
			}
		}
	}
	return float64(n) / float64(b.Dx()*b.Dy())
}

// A real arXiv vector figure rasterises to a non-blank image.
func TestRasterizeRealFigure(t *testing.T) {
	pdf, err := os.ReadFile("testdata/figure.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if frac := nonWhiteFraction(t, pdf); frac < 0.01 {
		t.Errorf("figure rendered nearly blank (%.3f non-white)", frac)
	}
}

func TestRasterizeErrors(t *testing.T) {
	if _, err := Rasterize([]byte("not a pdf"), 150); err == nil {
		t.Error("garbage bytes should fail to open")
	}
	pdf, err := os.ReadFile("testdata/figure.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RasterizePage(pdf, 0, 150); err == nil {
		t.Error("page 0 should be out of range")
	}
	if _, err := RasterizePage(pdf, 99, 150); err == nil {
		t.Error("page 99 should be out of range")
	}
}

// inkInLeftMargin is the share of non-white pixels in the leftmost twelfth of
// the page. On testdata/figure.pdf that band holds the eight y-axis category
// labels and nothing else: no rule, no marker, no plotting area. It is
// therefore a band that is inked if and only if the renderer drew text.
func inkInLeftMargin(t *testing.T, pdf []byte) float64 {
	t.Helper()
	img, err := Rasterize(pdf, 150)
	if err != nil {
		t.Fatalf("Rasterize: %v", err)
	}
	b := img.Bounds()
	edge := b.Min.X + b.Dx()*12/100
	var n, tot int
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < edge; x++ {
			tot++
			r, g, bl, _ := img.At(x, y).RGBA()
			if r < 0xf000 || g < 0xf000 || bl < 0xf000 {
				n++
			}
		}
	}
	if tot == 0 {
		t.Fatal("empty image")
	}
	return float64(n) / float64(tot)
}

// The text of a figure is drawn, not merely its curves.
//
// This is deliberately not [TestRasterizeRealFigure] with a higher threshold.
// A whole-page ink measure cannot see this defect: the renderer this package
// used until now drew every curve and marker of testdata/figure.pdf and not one
// character of its titles, axis labels, tick labels or legend, and it still
// inked 4.6% of the page — comfortably "non-blank". The loss only becomes
// visible in a band that holds text alone, where that renderer scored exactly
// zero pixels against the 4.0% measured here.
func TestRasterizeDrawsTheTextOfAFigure(t *testing.T) {
	pdf, err := os.ReadFile("testdata/figure.pdf")
	if err != nil {
		t.Fatal(err)
	}
	const want = 0.02 // half of what is measured, well clear of zero
	if got := inkInLeftMargin(t, pdf); got < want {
		t.Errorf("ink in the left margin = %.4f, want at least %.4f; the y-axis labels are missing", got, want)
	}
}
