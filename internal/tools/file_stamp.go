package tools

import (
	"crypto/sha256"
	"io"
	"os"
)

func stampFile(path string) (fileStamp, error) {
	file, err := os.Open(path)
	if err != nil {
		return fileStamp{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return fileStamp{}, err
	}
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return fileStamp{}, err
	}
	stamp := fileStamp{mod: info.ModTime(), size: info.Size()}
	copy(stamp.digest[:], digest.Sum(nil))
	return stamp, nil
}

func unchangedContent(path, before string, existed bool) func() error {
	return func() error {
		now, exists, err := currentContent(path)
		if err != nil {
			return err
		}
		if exists != existed || now != before {
			return staleFileError(path)
		}
		return nil
	}
}
