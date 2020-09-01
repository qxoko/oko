package main

import (
	"io"
	"os"
	"time"
	"strings"
	"io/ioutil"
	"path/filepath"
)

type File_Type int; const (
	STATIC File_Type = iota
	STATIC_PAGE // html files
	MARKUP
	DIR
)

type File_Action int; const (
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

	IsDraft bool

	Children []*File
	Page       *Page

	Mod time.Time
}

func walk(root string) (map[string]*File, time.Time) {
	var list = make(map[string]*File)
	var age time.Time

	if !path_exists(root) {
		return list, age // empty
	}

	var owner_directory *File

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		name   := info.Name()
		prefix := name[0:1]

		skip := false

		if (prefix == "." && len(path) > 1) || prefix == "_" {
			skip = true
		}

		// directories
		if info.IsDir() {
			if name == config.Output {
				skip = true
			}

			if name == root || strings.HasSuffix(root, name) {
				skip = false
			}

			if skip {
				return filepath.SkipDir
			}

			path, _ := filepath.Rel(root, path)

			file_info := &File {
				ID:   path,
				SourcePath: path,
				OutputPath: filepath.Join(config.Output, path),
				Type: DIR,
				Mod:  info.ModTime(),
			}

			owner_directory = file_info

			list[path] = file_info

			return nil
		}

		// files
		file_ext := filepath.Ext(name)

		for _, f := range config.Exclude {
			if file_ext == f {
				skip = true
				break
			}
		}

		if skip { return nil }

		file_type := STATIC

		path, _  = filepath.Rel(root, path)
		file_id := filepath.ToSlash(path)
		output  := filepath.Join(config.Output, file_id)

		is_any_special := true

		switch file_ext {
			case ".html":
				file_type = STATIC_PAGE
				config.PageCount++

			case ".ø", ".txt":
				if name == "robots.txt" {
					is_any_special = false
					break
				}
				file_type = MARKUP
				config.PageCount++

			default:
				is_any_special = false
		}

		if is_any_special {
			file_id = file_id[:len(path) - len(file_ext)]
			output  = filepath.Join(config.Output, file_id) + ".html"
		}

		t := info.ModTime()

		if t.After(age) {
			age = t // update youngest file
		}

		file_info := &File {
			ID:     file_id,
			SourcePath: path,
			OutputPath: output,
			Type:   file_type,
			Mod:    t,
		}

		owner_directory.Children = append(owner_directory.Children, file_info)

		list[file_id] = file_info

		return nil
	})

	return list, age
}

func support_files(age time.Time) map[string]bool {
	list := make(map[string]bool, 16)

	filepath.Walk("_data", func(path string, info os.FileInfo, err error) error {
		name        := info.Name()
		path_prefix := name[0:1]

		id_prefix := ""

		if info.IsDir() {
			switch name {
				case "plates":    id_prefix = "plate_"
				case "snippets":  id_prefix = "snip_"
				case "functions": id_prefix = "functions_"
				case "syntax":    id_prefix = "syntax_"
				default: return filepath.SkipDir
			}
		}

		if path_prefix == "." || path_prefix == "_" {
			return nil
		}

		if info.ModTime().After(age) {
			name = id_prefix + name[0:len(name) - len(filepath.Ext(name))]
			list[name] = true
		}

		return nil
	})

	return list
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
	os.MkdirAll(path, os.ModeDir)
}

func copy_file(src, dst string) {
	source, err := os.Open(src)

	if err != nil {
		panic(err)
	}

	defer source.Close()

	destination, err := os.OpenFile(dst, os.O_CREATE, 0755)

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