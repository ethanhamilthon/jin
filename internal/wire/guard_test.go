package wire

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var (
	tagPattern  = regexp.MustCompile(`(^|\n)[ \t]*</?[a-z]+(-[a-z]+)+[ >/\\]`)
	notePattern = regexp.MustCompile(`^\n?\[(exit code|command (timed|completed)|output (of|truncated)|truncated at|The current model)`)
	// Text of the commands in the system prompt file is allowed to say [command failed].
	allowedDirs = []string{"wire", "dyn"}
)

// TestNoMessageTextOutsideWire makes a new tag or note for the model show up
// here: the text must be written in this package, where a review sees all of it.
func TestNoMessageTextOutsideWire(t *testing.T) {
	root := filepath.Join("..")
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		for _, dir := range allowedDirs {
			if strings.HasPrefix(rel, dir+string(filepath.Separator)) {
				return nil
			}
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			if tagPattern.MatchString(text) || notePattern.MatchString(text) || hasTag(text) {
				t.Errorf("%s: message text outside the wire package: %.70q", fset.Position(lit.Pos()), text)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// hasTag finds a registered tag at the start of a literal or of a line.
func hasTag(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimLeft(line, " \t")
		for _, tag := range Tags {
			for _, form := range []string{"<" + tag, "</" + tag} {
				if rest, ok := strings.CutPrefix(line, form); ok && (rest == "" || strings.ContainsRune(" >/\\\"", rune(rest[0]))) {
					return true
				}
			}
		}
	}
	return false
}
