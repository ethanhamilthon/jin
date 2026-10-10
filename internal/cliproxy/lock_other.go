//go:build !unix

package cliproxy

import "errors"

func lock(string, string) (func(), error) {
	return nil, errors.New("managed CLIProxyAPI requires macOS or Linux")
}
