package ui

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
)

func testPicture(t *testing.T, w, h int) *provider.Image {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return &provider.Image{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(buf.Bytes())}
}

func TestPictureRowsDrawHalfBlocks(t *testing.T) {
	rows := entryRows(chatEntry{kind: core.UpdateInfo, text: "[image/png]", image: testPicture(t, 4, 4)}, 80)
	if len(rows) != 4 || len(rows[1].spans) != 4 || rows[1].spans[0].text != "▀" {
		t.Fatalf("rows %d, first picture row %+v", len(rows), rows[1])
	}
}

func TestPictureRowsStayInsideBounds(t *testing.T) {
	rows := entryRows(chatEntry{kind: core.UpdateInfo, text: "[big]", image: testPicture(t, 1000, 3000)}, 40)
	if len(rows) > maxImageRows+2 {
		t.Fatalf("%d rows", len(rows))
	}
	for _, row := range rows {
		if len(row.spans) > 38 {
			t.Fatalf("%d cells in a 40 wide chat", len(row.spans))
		}
	}
}
