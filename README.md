# go-tex/pdfrender

[![ci](https://github.com/go-tex/pdfrender/actions/workflows/ci.yml/badge.svg)](https://github.com/go-tex/pdfrender/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-tex/pdfrender.svg)](https://pkg.go.dev/github.com/go-tex/pdfrender)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD--3--Clause-blue.svg)](LICENSE)

Pure-Go (CGO=0) rasteriser for **PDF figures**, shaped to plug into
[go-tex/engine](https://github.com/go-tex/engine)'s `RasterizePDF` seam so
`\includegraphics` of a vector `.pdf` typesets as a real raster instead of a
placeholder. A thin wrapper over
[go-pdfkit/render](https://github.com/go-pdfkit/render).

## Why a separate module

A pure-Go PDF renderer is a heavy dependency. The engine core —
and its browser/wasm build — stay free of it: the engine exposes a `func` seam, and
a consumer that wants PDF figures (the CLI, loom) opts in with one line.

## Use

```go
import (
	"github.com/go-tex/engine"
	"github.com/go-tex/pdfrender"
)

func init() { engine.RasterizePDF = pdfrender.Rasterize } // \includegraphics{fig.pdf} now renders
```

Or rasterise directly:

```go
img, err := pdfrender.Rasterize(pdfBytes, 200) // first page at 200 DPI → image.Image
img, err  = pdfrender.RasterizePage(pdfBytes, 2, 300)
```

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright the go-tex/pdfrender authors.
