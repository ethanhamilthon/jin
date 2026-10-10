package cliproxy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"testing"
)

func archiveFixture(t *testing.T, headers []*tar.Header) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	writer := tar.NewWriter(gz)
	for _, header := range headers {
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Size == 3 {
			writer.Write([]byte("bin"))
		}
	}
	writer.Close()
	gz.Close()
	return buf.Bytes()
}

func TestArchiveAndChecksumValidation(t *testing.T) {
	valid := archiveFixture(t, []*tar.Header{{Name: "cli-proxy-api", Size: 3, Typeflag: tar.TypeReg}})
	hash := sha256.Sum256(valid)
	sums := []byte(fmt.Sprintf("%x  release.tar.gz\n", hash))
	if err := verify(valid, sums, "release.tar.gz"); err != nil {
		t.Fatal(err)
	}
	if err := verify([]byte("modified"), sums, "release.tar.gz"); err == nil {
		t.Fatal("bad checksum accepted")
	}
	binary, err := extract(valid)
	if err != nil || string(binary) != "bin" {
		t.Fatalf("extract=%q %v", binary, err)
	}
	for _, headers := range [][]*tar.Header{
		{{Name: "../cli-proxy-api", Size: 3, Typeflag: tar.TypeReg}},
		{{Name: "cli-proxy-api", Typeflag: tar.TypeSymlink, Linkname: "/bin/sh"}},
		{{Name: "cli-proxy-api", Size: 3, Typeflag: tar.TypeReg}, {Name: "cli-proxy-api", Size: 3, Typeflag: tar.TypeReg}},
		{{Name: "cli-proxy-api", Size: maxBinary + 1, Typeflag: tar.TypeReg}},
	} {
		if _, err := extract(archiveFixture(t, headers)); err == nil {
			t.Fatal("invalid archive accepted")
		}
	}
}

func TestAssetCompatibility(t *testing.T) {
	for _, version := range []string{"latest", "v9.0.0", "v8.0.22", "../../outside"} {
		if ValidVersion(version) {
			t.Fatal(version)
		}
	}
	for _, os := range []string{"darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			if _, err := asset(DefaultVersion, os, arch); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := asset(DefaultVersion, "windows", "arm64"); err == nil {
		t.Fatal("native Windows accepted")
	}
}
