package tools

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type grepFile struct{ abs, shown string }

var skippedDirs = []string{"node_modules", "vendor", "dist"}

// grepFiles lists the files to search under root, sorted by path. In a git
// repository that is what git does not ignore; elsewhere it is every file
// outside hidden folders and the usual dependency folders.
func grepFiles(ctx context.Context, root string, info os.FileInfo, args grepArgs) ([]grepFile, error) {
	if !info.IsDir() {
		return []grepFile{{root, args.Path}}, nil
	}
	rels, ok := gitFiles(ctx, root)
	if !ok {
		var err error
		if rels, err = walkFiles(ctx, root); err != nil {
			return nil, err
		}
	}
	var files []grepFile
	for _, rel := range rels {
		if !globMatch(args.Glob, rel) {
			continue
		}
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if st, err := os.Lstat(abs); err == nil && st.Mode().IsRegular() {
			files = append(files, grepFile{abs, filepath.Join(args.Path, filepath.FromSlash(rel))})
		}
	}
	return files, ctx.Err()
}

func gitFiles(ctx context.Context, root string) ([]string, bool) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, false
	}
	out, err := exec.CommandContext(ctx, "git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard").Output()
	if err != nil {
		return nil, false
	}
	var rels []string
	for _, rel := range bytes.Split(out, []byte{0}) {
		if len(rel) > 0 {
			rels = append(rels, string(rel))
		}
	}
	slices.Sort(rels)
	return slices.Compact(rels), true
}

func walkFiles(ctx context.Context, root string) ([]string, error) {
	var rels []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || ctx.Err() != nil {
			return filepath.SkipDir
		}
		name := d.Name()
		if d.IsDir() && p != root && (strings.HasPrefix(name, ".") || slices.Contains(skippedDirs, name)) {
			return filepath.SkipDir
		}
		if d.Type().IsRegular() {
			rel, _ := filepath.Rel(root, p)
			rels = append(rels, filepath.ToSlash(rel))
		}
		return nil
	})
	return rels, err
}
