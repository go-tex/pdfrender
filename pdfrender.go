// Copyright (c) the go-tex/pdfrender authors.
// SPDX-License-Identifier: BSD-3-Clause

// Package pdfrender rasterises a PDF page to an image in pure Go (CGO=0). It is a
// thin wrapper over github.com/go-pdfkit/render, shaped to plug straight into
// go-tex/engine's RasterizePDF seam:
//
//	engine.RasterizePDF = pdfrender.Rasterize
//
// so \includegraphics of a vector .pdf figure typesets as a real raster. The engine
// core stays free of this (heavy) renderer; a consumer that wants PDF figures — the
// CLI, loom — opts in with the line above, while the browser/wasm build leaves the
// seam nil and shows a placeholder.
package pdfrender

import (
	"fmt"
	"image"

	"github.com/go-gfx/gfx/raster"
	"github.com/go-pdfkit/reader"
	"github.com/go-pdfkit/render"
)

// Rasterize renders the first page of a PDF document to an image at the given DPI.
// It matches the signature of go-tex/engine's RasterizePDF hook.
func Rasterize(pdf []byte, dpi float64) (image.Image, error) {
	return RasterizePage(pdf, 1, dpi)
}

// NumPages reports how many pages a PDF document holds, or 0 when it cannot be
// opened. go-tex/engine's \includepdf needs it to resolve pdfpages' ranges —
// pages={2-5} and pages=- name a count the document alone knows — and counting
// "/Type /Page" in the bytes does not work on a PDF whose page tree lives in an
// object stream, which most modern writers produce.
func NumPages(pdf []byte) int {
	d, err := reader.Open(pdf)
	if err != nil {
		return 0
	}
	return d.PageCount()
}

// RasterizePage renders a specific 1-based page of a PDF document at the given DPI.
func RasterizePage(pdf []byte, page int, dpi float64) (image.Image, error) {
	d, err := reader.Open(pdf)
	if err != nil {
		return nil, fmt.Errorf("pdfrender: open: %w", err)
	}
	if n := d.PageCount(); page < 1 || page > n {
		return nil, fmt.Errorf("pdfrender: page %d out of range (1..%d)", page, n)
	}
	img, err := renderFitting(d, page, dpi)
	if err != nil {
		return nil, fmt.Errorf("pdfrender: render page %d: %w", page, err)
	}
	// raster.Image's At returns a color.RGBA, not a color.Color, so it is not
	// an image.Image; NRGBA is what a caller encoding or compositing expects.
	return img.ToNRGBA(), nil
}

// halvings is how many times renderFitting will halve the resolution before it
// gives up. Four takes 150 dpi down to about 9, which is past any figure that
// could be worth drawing, and 1/256th of the pixels of the first attempt.
const halvings = 4

// renderFitting draws a page at dpi, halving the resolution and trying again
// while the renderer refuses the size.
//
// render.Page bounds what it will allocate: past forty megapixels it returns an
// error rather than a very large image. That bound is written for a PAGE, and a
// figure is not one. One of the 770 figures in the go-tex corpus is 13580 by
// 8153 pixels at 150 dpi — a hundred and ten megapixels, 443 MB of RGBA before
// the NRGBA copy — and it is a figure that will be scaled into a text column,
// where a tenth of that resolution is already more than the page can show.
//
// The refusal is made from the page box alone, before anything is drawn, so a
// refused attempt costs nothing and there is no need to guess the size first.
func renderFitting(d *reader.Document, page int, dpi float64) (*raster.Image, error) {
	var err error
	for i := 0; i <= halvings; i++ {
		var img *raster.Image
		img, err = render.Page(d, page, render.Options{DPI: dpi})
		if err == nil {
			return img, nil
		}
		dpi /= 2
	}
	return nil, err
}
