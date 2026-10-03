package provider

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

func readSSE(r io.Reader, handle func([]byte) (bool, error)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64<<10), maxProviderBody)
	var data []byte
	dispatch := func() (bool, error) {
		if len(data) == 0 {
			return false, nil
		}
		stop, err := handle(data)
		data = nil
		return stop, err
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if stop, err := dispatch(); stop || err != nil {
				return err
			}
			continue
		}
		if value, ok := strings.CutPrefix(line, "data:"); ok {
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, strings.TrimPrefix(value, " ")...)
			if len(data) > maxProviderBody {
				return errors.New("provider SSE event exceeds 4 MiB")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return transient(errors.New("stream read failed: " + err.Error()))
	}
	_, err := dispatch()
	return err
}
