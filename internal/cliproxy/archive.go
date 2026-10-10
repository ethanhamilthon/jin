package cliproxy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
)

func extract(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	var binary []byte
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.New("corrupt CLIProxyAPI archive")
		}
		if header.Name != "cli-proxy-api" {
			continue
		}
		if binary != nil || (header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA) || header.Size <= 0 || header.Size > maxBinary {
			return nil, errors.New("invalid CLIProxyAPI binary entry")
		}
		binary, err = io.ReadAll(io.LimitReader(reader, maxBinary+1))
		if err != nil || int64(len(binary)) != header.Size {
			return nil, errors.New("invalid CLIProxyAPI binary size")
		}
	}
	if binary == nil {
		return nil, errors.New("CLIProxyAPI binary was not found in the archive")
	}
	return binary, nil
}
