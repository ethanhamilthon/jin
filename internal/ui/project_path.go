package ui

import "jin/internal/session"

func projectPath(path, baseDir string) (string, error) { return session.ProjectPath(path, baseDir) }
