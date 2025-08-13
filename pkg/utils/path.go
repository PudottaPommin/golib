package utils

import (
	"path"
	"path/filepath"
)

func PathJoinRel(elements ...string) string {
	parts := make([]string, len(elements))
	for i, e := range elements {
		if e == "" {
			continue
		}
		parts[i] = path.Clean("/" + e)
	}
	p := path.Join(parts...)
	switch p {
	case "":
		return ""
	case "/":
		return "."
	}
	return p[1:]
}

func PathJoinRelX(elements ...string) string {
	parts := make([]string, 0, len(elements))
	for i := range elements {
		if elements[i] == "" {
			continue
		}
		parts = append(parts, path.Clean("/"+filepath.ToSlash(elements[i])))
	}
	return PathJoinRel(parts...)
}

func FilePathJoinAbs(base string, sub ...string) string {
	elements := make([]string, 1, len(sub)+1)
	elements[0] = filepath.Clean(filepath.ToSlash(base))
	for i := range sub {
		elements = append(elements, filepath.Clean(filepath.ToSlash(sub[i])))
	}
	return filepath.ToSlash(filepath.Join(elements...))
}
