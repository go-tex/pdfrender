// Copyright (c) the go-tex/pdfrender authors.
// SPDX-License-Identifier: BSD-3-Clause

package pdfrender

import (
	"testing"

	"github.com/go-pdfkit/reader"
)

// hugePage builds a one-page document whose media box is `side` points square,
// with a filled square covering the lower-left quarter of it.
func hugePage(t *testing.T, side float64) []byte {
	t.Helper()
	w := reader.NewWriter("1.7")
	pagesRef := w.Reserve()
	content := "0 0 0 rg 0 0 " + ftoa(side/2) + " " + ftoa(side/2) + " re f"
	page := reader.Dict{
		"Type":     reader.Name("Page"),
		"Parent":   pagesRef,
		"Contents": w.Add(&reader.Stream{Dict: reader.Dict{}, Raw: []byte(content)}),
		"MediaBox": reader.Array{reader.Real(0), reader.Real(0), reader.Real(side), reader.Real(side)},
	}
	pageRef := w.Add(page)
	w.Put(pagesRef, reader.Dict{"Type": reader.Name("Pages"),
		"Kids": reader.Array{pageRef}, "Count": reader.Integer(1)})
	root := w.Add(reader.Dict{"Type": reader.Name("Catalog"), "Pages": pagesRef})
	out, err := w.Finish(reader.Dict{"Root": root})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func ftoa(f float64) string {
	s := ""
	n := int(f)
	if n == 0 {
		return "0"
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// A figure too large to draw at the asked-for resolution is drawn at a smaller
// one rather than refused.
//
// render.Page allocates nothing past forty megapixels; it answers with an error
// made from the page box, before drawing. That bound is written for a page, and
// the go-tex corpus holds a figure that is 13580 by 8153 pixels at 150 dpi.
// Refusing it would frame a placeholder where the paper has a picture.
func TestAFigureTooLargeToDrawAtOnceIsDrawnSmaller(t *testing.T) {
	// 20000 points square is 41666 by 41666 pixels at 150 dpi, about 1.7
	// gigapixels: forty times over the limit, so it takes three halvings.
	pdf := hugePage(t, 20000)

	img, err := Rasterize(pdf, 150)
	if err != nil {
		t.Fatalf("Rasterize: %v", err)
	}
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		t.Fatal("empty image")
	}
	if got := b.Dx() * b.Dy(); got > 40<<20 {
		t.Errorf("image is %d pixels, which is past the limit the renderer refuses at", got)
	}
	// The lower-left quarter is filled, so about a quarter of the page is inked.
	// Anything near zero would mean a blank image of the right size.
	if got := nonWhiteFraction(t, pdf); got < 0.2 {
		t.Errorf("ink = %.4f, want about 0.25; the page came out blank", got)
	}
}
