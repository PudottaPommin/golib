package assetsfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/pudottapommin/golib/pkg/set"
	"github.com/pudottapommin/golib/pkg/utils"
)

type Layer struct {
	name      string
	fs        fs.FS
	localPath string
}

func (l *Layer) Name() string {
	return l.name
}

func (l *Layer) Open(name string) (fs.File, error) {
	return l.fs.Open(name)
}

func (l *Layer) ReadDir(name string) ([]fs.DirEntry, error) {
	dirEntries, err := fs.ReadDir(l.fs, name)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	return dirEntries, err
}

func Local(name, base string, sub ...string) *Layer {
	base, err := filepath.Abs(base)
	if err != nil {
		panic(err)
	}
	root := utils.FilePathJoinAbs(base, sub...)
	return &Layer{name: name, fs: os.DirFS(root), localPath: root}
}

func Blobs(name string, fs fs.FS) *Layer {
	return &Layer{name: name, fs: fs}
}

type LayeredFS struct {
	layers []*Layer
}

func NewLayered(layers ...*Layer) *LayeredFS {
	return &LayeredFS{layers: layers}
}

func (l *LayeredFS) Open(name string) (fs.File, error) {
	for _, layer := range l.layers {
		f, err := layer.Open(name)
		if err == nil || !os.IsNotExist(err) {
			return f, err
		}
	}
	return nil, fs.ErrNotExist
}

func (l *LayeredFS) ReadFile(elements ...string) ([]byte, error) {
	b, _, err := l.LayeredReadFile(elements...)
	return b, err
}

func (l *LayeredFS) LayeredReadFile(elements ...string) ([]byte, string, error) {
	name := utils.PathJoinRel(elements...)
	for _, layer := range l.layers {
		f, err := layer.Open(name)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, layer.name, err
		}
		b, err := io.ReadAll(f)
		_ = f.Close()
		return b, layer.name, err
	}
	return nil, "", fs.ErrNotExist
}

func (l *LayeredFS) ListFiles(name string, fileInclude FileInclude) ([]string, error) {
	fileSet := make(set.Set[string])
	for _, layer := range l.layers {
		files, err := layer.ReadDir(name)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if shouldInclude(file, fileInclude) {
				fileSet.Add(file.Name())
			}
		}
	}
	files := slices.Sorted(fileSet.Seq())
	return files, nil
}

func (l *LayeredFS) ListAllFiles(name string, fileInclude FileInclude) ([]string, error) {
	return listAllFiles(l.layers, name, fileInclude)
}

func listAllFiles(layers []*Layer, name string, fileInclude FileInclude) ([]string, error) {
	fileSet := make(set.Set[string])
	var listFn func(dir string) error
	listFn = func(dir string) error {
		for _, layer := range layers {
			files, err := layer.ReadDir(dir)
			if err != nil {
				return err
			}
			for _, file := range files {
				path := utils.PathJoinRelX(dir, file.Name())
				if shouldInclude(file, fileInclude) {
					fileSet.Add(path)
				}
				if file.IsDir() {
					if err = listFn(path); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	if err := listFn(name); err != nil {
		return nil, err
	}
	files := slices.Sorted(fileSet.Seq())
	return files, nil
}

func shouldInclude(d fs.DirEntry, fileInclude FileInclude) bool {
	if d.Name() == "." {
		return false
	}
	if fileInclude == FileIncludeDirs && !d.IsDir() {
		return false
	}
	if fileInclude == FileIncludeFiles && d.IsDir() {
		return false
	}
	return true
}

type FileInclude uint8

const (
	FileIncludeAll FileInclude = iota
	FileIncludeFiles
	FileIncludeDirs
)
