package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

func verify(sums []byte, name string, data []byte) error {
	sum := sha256.Sum256(data)
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			if fields[0] != hex.EncodeToString(sum[:]) {
				return errors.New("checksum mismatch for " + name)
			}
			return nil
		}
	}
	return errors.New("no checksum listed for " + name)
}

func extract(archive []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err != nil {
			return nil, errors.New("the archive has no jin binary")
		}
		if header.Name != "jin" {
			continue
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			return nil, errors.New("jin entry is not a regular file")
		}
		if header.Size < 0 || header.Size > maxArchive {
			return nil, errors.New("binary size exceeds limit")
		}
		data, err := io.ReadAll(io.LimitReader(reader, maxArchive+1))
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxArchive {
			return nil, errors.New("binary exceeds maximum allowed size")
		}
		if int64(len(data)) != header.Size {
			return nil, errors.New("binary size mismatch")
		}
		return data, nil
	}
}
