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
	img, err := render.Page(d, page, render.Options{DPI: dpi})
	if err != nil {
		return nil, fmt.Errorf("pdfrender: render page %d: %w", page, err)
	}
	// raster.Image's At returns a color.RGBA, not a color.Color, so it is not
	// an image.Image; NRGBA is what a caller encoding or compositing expects.
	return img.ToNRGBA(), nil
}
