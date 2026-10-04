package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
)

const (
	maxReadOutput = 32 << 10
	binarySniff   = 8 << 10
)

const readSchema = `{"type":"function","function":{"name":"read","description":"Read a file. Each line comes back as ` + "`<line number><TAB><text>`" + `; the number and tab are not part of the file. At most 32 KB per call: use offset and limit (1-based) for big files. Binary files are refused. Pictures (png, jpeg, gif, webp, bmp) are attached for you to see.","parameters":{"type":"object","properties":{"path":{"type":"string","description":"File path to read"},"offset":{"type":"integer","minimum":1,"description":"1-based line number to start from"},"limit":{"type":"integer","minimum":1,"description":"Maximum number of lines to return"}},"required":["path"],"additionalProperties":false}}}`

type Read struct{ seen *Seen }

func NewRead() Read { return Read{} }

// NewReadSeen shares seen with the other file tools of a session.
func NewReadSeen(seen *Seen) Read { return Read{seen: seen} }

func (Read) Name() string { return "read" }

func (Read) Schema() json.RawMessage { return json.RawMessage(readSchema) }

func (Read) Summary(argumentsJSON string) (string, bool) {
	args, err := parseReadArgs(argumentsJSON)
	if err != nil {
		return "", false
	}
	return args.summary(), true
}

func (r Read) Run(ctx context.Context, argumentsJSON string) (string, error) {
	text, _, err := r.RunImages(ctx, argumentsJSON)
	return text, err
}

// RunImages reads a picture as an attachment and anything else as text.
func (r Read) RunImages(ctx context.Context, argumentsJSON string) (string, []Image, error) {
	args, err := parseReadArgs(argumentsJSON)
	if err != nil {
		return "", nil, err
	}
	if info, err := os.Stat(args.Path); err != nil {
		return "", nil, err
	} else if !info.Mode().IsRegular() {
		return "", nil, errors.New(args.Path + " is not a regular file")
	}
	file, err := os.Open(args.Path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", nil, err
	}
	if !info.Mode().IsRegular() {
		return "", nil, errors.New(args.Path + " is not a regular file")
	}
	r.seen.Remember(args.Path)
	head := make([]byte, binarySniff)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", nil, err
	}
	if bytes.IndexByte(head[:n], 0) >= 0 || hasImageHeader(head[:n]) {
		return readPicture(file, info.Size(), args.Path, hasImageHeader(head[:n]))
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", nil, err
	}
	text, err := readLines(ctx, file, args)
	return text, nil, err
}

func readPicture(file *os.File, size int64, path string, knownImage bool) (string, []Image, error) {
	refusal := errors.New(path + " is a binary file; read only handles text and pictures")
	if size > maxImageFile {
		if knownImage {
			return "", nil, fmt.Errorf("%s is an image file of %d MB; the limit is %d MB", path, size>>20, maxImageFile>>20)
		}
		return "", nil, refusal
	}
	data, err := io.ReadAll(io.NewSectionReader(file, 0, size))
	if err != nil {
		return "", nil, err
	}
	picture, err := loadImage(data)
	switch {
	case err == nil:
		return "Read image file [" + picture.MimeType + "]", []Image{picture}, nil
	case errors.Is(err, errNotImage):
		return "", nil, refusal
	}
	return "", nil, errors.New(path + ": " + err.Error())
}

func hasImageHeader(head []byte) bool {
	_, _, err := image.DecodeConfig(bytes.NewReader(head))
	return err == nil
}
