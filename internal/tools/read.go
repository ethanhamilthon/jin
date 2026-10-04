package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"io"
	"os"
)

const (
	maxReadOutput = 32 << 10
	binarySniff   = 8 << 10
	maxImageFile  = 64 << 20
)

const readSchema = `{"type":"function","function":{"name":"read","description":"Read a file from disk, optionally a line range. Pictures (png, jpeg, gif, webp, bmp) are attached for you to see","parameters":{"type":"object","properties":{"path":{"type":"string","description":"File path to read"},"offset":{"type":"integer","description":"1-based line number to start from"},"limit":{"type":"integer","description":"Maximum number of lines to return"}},"required":["path"],"additionalProperties":false}}}`

type Read struct{ seen *Seen }

func NewRead() Read { return Read{} }

// NewReadSeen shares seen with the other file tools of a session.
func NewReadSeen(seen *Seen) Read { return Read{seen: seen} }

func (Read) Name() string { return "read" }

func (Read) Schema() json.RawMessage { return json.RawMessage(readSchema) }

func (Read) Summary(argumentsJSON string) (string, bool) {
	args, ok := parseReadArgs(argumentsJSON)
	if !ok {
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
	args, ok := parseReadArgs(argumentsJSON)
	if !ok {
		return "", nil, errors.New("invalid read tool arguments")
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
		return readPicture(file, info.Size(), args.Path)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", nil, err
	}
	text, err := readLines(ctx, file, args)
	return text, nil, err
}

func readPicture(file *os.File, size int64, path string) (string, []Image, error) {
	refusal := errors.New(path + " is a binary file; read only handles text and pictures")
	if size > maxImageFile {
		return "", nil, refusal
	}
	data, err := io.ReadAll(io.NewSectionReader(file, 0, size))
	if err != nil {
		return "", nil, err
	}
	if picture, ok := loadImage(data); ok {
		return "Read image file [" + picture.MimeType + "]", []Image{picture}, nil
	}
	return "", nil, refusal
}

func hasImageHeader(head []byte) bool {
	_, _, err := image.DecodeConfig(bytes.NewReader(head))
	return err == nil
}
