package tools

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"

	xdraw "golang.org/x/image/draw"
)

func encodePNG(img image.Image) (string, bool) {
	var buf bytes.Buffer
	if png.Encode(&buf, img) != nil {
		return "", false
	}
	return finish(buf.Bytes())
}

// encodeJPEG flattens transparency onto white first.
func encodeJPEG(img image.Image, quality int) (string, bool) {
	flat := image.NewRGBA(img.Bounds())
	xdraw.Draw(flat, flat.Bounds(), image.NewUniform(color.White), image.Point{}, xdraw.Src)
	xdraw.Draw(flat, flat.Bounds(), img, img.Bounds().Min, xdraw.Over)
	var buf bytes.Buffer
	if jpeg.Encode(&buf, flat, &jpeg.Options{Quality: quality}) != nil {
		return "", false
	}
	return finish(buf.Bytes())
}

func finish(data []byte) (string, bool) {
	if base64.StdEncoding.EncodedLen(len(data)) > maxImageBase64 {
		return "", false
	}
	return base64.StdEncoding.EncodeToString(data), true
}
