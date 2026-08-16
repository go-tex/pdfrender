// Copyright (c) the go-tex/pdfrender authors.
// SPDX-License-Identifier: BSD-3-Clause

// Package pdfrender rasterises a PDF page to an image in pure Go (CGO=0). It is a
// thin wrapper over the reference renderer github.com/ajroetker/pdf/render, shaped
// to plug straight into go-tex/engine's RasterizePDF seam:
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

	"github.com/ajroetker/pdf/render"
)

// Rasterize renders the first page of a PDF document to an image at the given DPI.
// It matches the signature of go-tex/engine's RasterizePDF hook.
func Rasterize(pdf []byte, dpi float64) (image.Image, error) {
	return RasterizePage(pdf, 1, dpi)
}

// RasterizePage renders a specific 1-based page of a PDF document at the given DPI.
func RasterizePage(pdf []byte, page int, dpi float64) (image.Image, error) {
	r, err := render.NewRenderer(pdf)
	if err != nil {
		return nil, fmt.Errorf("pdfrender: open: %w", err)
	}
	defer r.Close()
	if page < 1 || page > r.NumPages() {
		return nil, fmt.Errorf("pdfrender: page %d out of range (1..%d)", page, r.NumPages())
	}
	img, err := r.RenderPage(page, dpi)
	if err != nil {
		return nil, fmt.Errorf("pdfrender: render page %d: %w", page, err)
	}
	return img, nil
}
