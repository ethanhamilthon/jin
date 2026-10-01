package tools

import (
	"bytes"
	"encoding/base64"
	"image"

	_ "image/gif"

	_ "golang.org/x/image/bmp"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	maxImageSide   = 2000
	maxImageBase64 = 4500 << 10
)

// Image is a picture a tool hands to the model.
type Image struct {
	MimeType string
	Data     string
}

// loadImage returns the picture in data, shrunk to fit the provider limits,
// or ok=false when data is not an image.
func loadImage(data []byte) (Image, bool) {
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Image{}, false
	}
	sendable := format == "png" || format == "jpeg" || format == "gif" || format == "webp"
	if sendable && config.Width <= maxImageSide && config.Height <= maxImageSide && base64.StdEncoding.EncodedLen(len(data)) <= maxImageBase64 {
		return Image{MimeType: "image/" + format, Data: base64.StdEncoding.EncodeToString(data)}, true
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Image{}, false
	}
	return shrink(decoded)
}

// shrink scales to the side limit, then keeps lowering JPEG quality and size
// until the encoded picture fits.
func shrink(src image.Image) (Image, bool) {
	side := maxImageSide
	for side >= 64 {
		scaled := fit(src, side)
		if encoded, ok := encodePNG(scaled); ok {
			return Image{MimeType: "image/png", Data: encoded}, true
		}
		for _, quality := range []int{80, 60, 40} {
			if encoded, ok := encodeJPEG(scaled, quality); ok {
				return Image{MimeType: "image/jpeg", Data: encoded}, true
			}
		}
		side = side * 3 / 4
	}
	return Image{}, false
}

func fit(src image.Image, side int) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= side && h <= side {
		return src
	}
	if w >= h {
		h, w = max(1, h*side/w), side
	} else {
		w, h = max(1, w*side/h), side
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)
	return dst
}
