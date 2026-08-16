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
