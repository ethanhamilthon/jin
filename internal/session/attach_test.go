package session

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writePNG(t *testing.T) string {
	path := filepath.Join(t.TempDir(), "a.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAttachImagesSendsPicturesApart(t *testing.T) {
	images := []Image{{Label: "[image 01]", Path: writePNG(t)}, {Label: "[image 02]", Path: writePNG(t)}}
	clean, shown, loaded := attachImages("see [image 01]", "see [image 01]", images, true)
	if clean != "see" || shown != "see [image 01]\n[image 02]" || len(loaded) != 2 {
		t.Fatalf("clean %q, shown %q, loaded %d", clean, shown, len(loaded))
	}
}

func TestAttachImagesKeepsPathOfBrokenPicture(t *testing.T) {
	clean, shown, loaded := attachImages("", "", []Image{{Label: "[image 01]", Path: "/missing.png"}}, true)
	if clean != "[image 01: /missing.png]" || shown != "[image 01]" || len(loaded) != 0 {
		t.Fatalf("clean %q, shown %q, loaded %d", clean, shown, len(loaded))
	}
}

func TestAttachImagesSkipsPicturesForModelWithoutVision(t *testing.T) {
	_, _, loaded := attachImages("", "", []Image{{Label: "[image 01]", Path: writePNG(t)}}, false)
	if len(loaded) != 0 {
		t.Fatalf("loaded %d", len(loaded))
	}
}
