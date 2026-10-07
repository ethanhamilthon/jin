package ui

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"sync"

	"github.com/gdamore/tcell/v3"
	tcolor "github.com/gdamore/tcell/v3/color"
	xdraw "golang.org/x/image/draw"

	"jin/internal/core"
	"jin/internal/provider"
)

const (
	maxImageCols = 64
	maxImageRows = 16
)

var decoded sync.Map

func decodePicture(picture *provider.Image) (image.Image, bool) {
	if cached, ok := decoded.Load(picture); ok {
		return cached.(image.Image), true
	}
	data, err := base64.StdEncoding.DecodeString(picture.Data)
	if err != nil {
		return nil, false
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, false
	}
	decoded.Store(picture, img)
	return img, true
}

// pictureRows shows a picture under its label, two pixels per cell: the
// upper half block takes the top pixel as foreground and the lower as background.
func pictureRows(entry chatEntry, width int) []chatRow {
	rows := []chatRow{{kind: core.UpdateInfo, text: entry.text}}
	img, ok := decodePicture(entry.image)
	if !ok {
		return append(rows, chatRow{})
	}
	pixels := scaleToCells(img, min(width-2, maxImageCols), maxImageRows*2)
	for y := 0; y < pixels.Bounds().Dy(); y += 2 {
		var spans []chatSpan
		for x := 0; x < pixels.Bounds().Dx(); x++ {
			style := tcell.StyleDefault.Foreground(cellColor(pixels.At(x, y))).Background(cellColor(pixels.At(x, y+1)))
			spans = append(spans, chatSpan{text: "▀", style: style})
		}
		rows = append(rows, chatRow{kind: core.UpdateInfo, spans: spans})
	}
	return append(rows, chatRow{})
}

func scaleToCells(img image.Image, cols, pixelRows int) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	outW := max(1, min(cols, w))
	outH := max(2, h*outW/w)
	if outH > pixelRows {
		outH = pixelRows
		outW = max(1, w*outH/h)
	}
	outH += outH % 2
	out := image.NewRGBA(image.Rect(0, 0, outW, outH))
	draw.Draw(out, out.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(out, out.Bounds(), img, bounds, xdraw.Over, nil)
	return out
}

func cellColor(c color.Color) tcolor.Color {
	r, g, b, _ := c.RGBA()
	return tcolor.NewRGBColor(int32(r>>8), int32(g>>8), int32(b>>8))
}
