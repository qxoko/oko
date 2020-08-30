package main

import (
	"io"
	"os"
	"time"
	"strings"
	"io/ioutil"
	"path/filepath"
)

type File_Type int

const (
	STATIC File_Type = iota
	MARKUP
	DIR
)

type File_Action int

const (
	NONE File_Action = iota
	NEEDS_UPDATE
	NEEDS_DELETE
)

type File struct {
	ID         string
	SourcePath string
	OutputPath string

	Type   File_Type
	Action File_Action

	Page *Page

	Mod    time.Time
}

func walk(root string) (map[string]*File, time.Time) {
	var list = make(map[string]*File)
	var age time.Time

	if !path_exists(root) {
		return list, age // empty
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		name   := info.Name()
		prefix := name[0:1]

		skip := false

		if (prefix == "." && len(path) > 1) || prefix == "_" {
			skip = true
		}

		if !info.IsDir() {
			if skip {
				return nil
			}

			file_ext    := filepath.Ext(name)
			file_format := ANY

			path, _ := filepath.Rel(root, path)
			file_id := filepath.ToSlash(path)

			dir := filepath.Dir(path)

			is_any_special := false

			switch e := file_ext {
				case ".html":
					file_format    = HTML
					is_any_special = true

				case ".ø":
					file_format    = OKO
					is_any_special = true

				case ".txt":
					if name == "robots.txt" {
						return nil
					}
					file_format    = OKO
					is_any_special = true
			}

			if is_any_special {
				file_id = fild_id[:len(path) - len(file_ext)]
			}

			t := info.ModTime()

			if t.After(age) {
				age = t // update youngest file
			}

			file_info := &File{
				ID:     file_id,
				Path:   path,
				Dir:    dir,
				Format: file_format,
				Mod:    t,
			}

			list[file_id] = file_info

		} else {
			if name == config.Output {
				skip = true
			}

			if name == root || strings.HasSuffix(root, name) {
				skip = false
			}

			if skip {
				return filepath.SkipDir
			}
		}
		return nil
	})

	return list, age
}

//
// file ops
//
func path_exists(path string) bool {
	stat, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return stat.IsDir()
}

func file_exists(path string) bool {
	stat, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return !stat.IsDir()
}

func file_data(path string) (os.FileInfo, bool) {
	stat, err := os.Stat(path)

	if os.IsNotExist(err) {
		return nil, false
	}

	return stat, true
}

func mkdir(path string) {
	os.MkdirAll(path, os.ModePerm)
}

func copy_file(src, dst string) {
	source, err := os.Open(src)

	if err != nil {
		panic(err)
	}

	defer source.Close()

	destination, err := os.Create(dst)

	if err != nil {
		panic(err)
	}

	defer destination.Close()

	_, err = io.Copy(destination, source)

	if err != nil {
		panic(err)
	}
}

func delete_file(path string) {
	err := os.RemoveAll(path)

	if err != nil {
		panic(err)
	}
}

func load_file_bytes(path string) []byte {
	file, err := os.Open(path)

	defer file.Close()

	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(file)

	if err != nil {
		panic(err)
	}

	return b
}
