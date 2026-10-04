package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

func archiveEntry(header *tar.Header, body []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(header)
	if len(body) > 0 {
		_, _ = tw.Write(body)
	}
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestExtractValidation(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr string
	}{
		{
			name: "symlink jin",
			data: archiveEntry(&tar.Header{
				Name:     "jin",
				Typeflag: tar.TypeSymlink,
				Linkname: "/bin/sh",
			}, nil),
			wantErr: "not a regular file",
		},
		{
			name: "header size exceeds limit",
			data: archiveEntry(&tar.Header{
				Name:     "jin",
				Typeflag: tar.TypeReg,
				Size:     maxArchive + 1,
			}, nil),
			wantErr: "binary size exceeds limit",
		},
		{
			name: "no jin binary",
			data: archiveEntry(&tar.Header{
				Name:     "other",
				Typeflag: tar.TypeReg,
				Size:     4,
			}, []byte("test")),
			wantErr: "has no jin binary",
		},
		{
			name:    "corrupt gzip",
			data:    []byte("not a gzip"),
			wantErr: "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extract(tt.data)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("extract() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestExtractValid(t *testing.T) {
	content := []byte("#!/bin/sh\necho test\n")
	archive := archiveEntry(&tar.Header{
		Name:     "jin",
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
	}, content)
	got, err := extract(archive)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("extract got %q, want %q", got, content)
	}
}
