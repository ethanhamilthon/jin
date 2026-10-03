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
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err != nil {
			return nil, errors.New("the archive has no jin binary")
		}
		if header.Name == "jin" && header.Typeflag == tar.TypeReg {
			return io.ReadAll(io.LimitReader(reader, maxArchive))
		}
	}
}
