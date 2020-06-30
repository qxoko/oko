package main

import (
	"io"
	"os"
	"time"
	"strings"
	"io/ioutil"
	"path/filepath"
)

var PageList = make(map[string]*Page)

type File_Format int

const (
	ANY File_Format = iota
	OKO
	HTML
)

type File_Info struct {
	ID   string
	Path string
	Dir  string

	Exclude bool

	Format File_Format
	Mod    time.Time
}

type Page struct {
	ID         string
	SourcePath string
	OutputPath string
	URLPath    string

	Style      []string
	Script     []string

	CurrentParent *Page // @hack

	IsDraft    bool
	IsIndex    bool
	Format     File_Format

	Plate      *Plate
	List       *Token_List

	Vars       map[string]string
	Meta       map[string]string
	Tags       map[string]bool
}

func make_page(info *File_Info) *Page {
	if p, exists := PageList[info.ID]; exists {
		return p
	}

	output := filepath.Join(config.Output, info.ID + ".html")

	new_page := &Page{
		ID: info.ID,
		SourcePath: info.Path,
		OutputPath: output,
		Format: info.Format,
	}

	new_page.Vars = make(map[string]string, 8)
	new_page.Meta = make(map[string]string, 8)
	new_page.Tags = make(map[string]bool,   8)

	if info.ID == "index" {
		new_page.URLPath = ""
		new_page.IsIndex = true
	} else {
		new_page.URLPath = "/" + strings.Replace(info.ID, "/index", "", 1)
		new_page.IsIndex = false
	}

	new_page.Vars["page_path"] = new_page.URLPath

	PageList[info.ID] = new_page

	return new_page
}

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

func walk(root string, extensions ...string) (map[string]*File_Info, time.Time) {
	check_ext := false

	if len(extensions) > 0 {
		check_ext = true
	}

	var list = make(map[string]*File_Info)
	var youngest time.Time

	if !path_exists(root) {
		return list, youngest
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		name   := info.Name()
		prefix := name[0:1]

		skip := false

		if prefix == "." && len(path) > 1 || prefix == "_" {
			skip = true
		}

		if !info.IsDir() {
			if skip {
				return nil
			}

			file_ext    := filepath.Ext(name)
			file_format := ANY

			path, _ := filepath.Rel(root, path)
			file_id := filepath.ToSlash(path) // id includes extension

			dir := filepath.Dir(path)

			if check_ext {
				is_match := false

				// strip extension for markup files
				file_id = file_id[:len(path) - len(file_ext)]

				if file_ext == ".html" {
					file_format = HTML
				} else {
					file_format = OKO
				}

				if file_ext == ".txt" {
					if name == "robots.txt" {
						return nil
					}
				}

				for _, e := range extensions {
					if e == file_ext {
						is_match = true
						break
					}
				}

				if !is_match {
					return nil
				}
			}

			t := info.ModTime()

			if t.After(youngest) {
				youngest = t
			}

			file_info := &File_Info{
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

	return list, youngest
}

func compare_files(source, output map[string]*File_Info) (map[string]*File_Info, map[string]*File_Info) {
	cap := len(source)

	mod := make(map[string]*File_Info, cap)
	del := make(map[string]*File_Info, cap)

	for _, src := range source {
		if dst, ok := output[src.ID]; ok {
			if src.Mod.After(dst.Mod) {
				mod[src.ID] = src
			}
		} else {
			mod[src.ID] = src
		}
	}

	for _, src := range output {
		if f, ok := source[src.ID]; !ok {
			del[src.ID] = src
		} else if f.Exclude {
			del[src.ID] = src
		}
	}

	return mod, del
}

func compare_dirs(source, output, file_mod, file_del map[string]*File_Info) (map[string]bool, map[string]bool) {
	cap := len(source)

	source_dirs := make(map[string]bool, cap)
	output_dirs := make(map[string]bool, cap)

	for _, f := range source {
		if f.Dir == "." { continue }
		source_dirs[f.Dir] = true
	}
	for _, f := range output {
		if f.Dir == "." { continue }
		output_dirs[f.Dir] = true
	}

	mod := make(map[string]bool, cap)
	del := make(map[string]bool, cap)

	for _, f := range source {
		if f.Dir == "." { continue }
		if !output_dirs[f.Dir] {
			mod[f.Dir] = true
		}
	}
	for _, f := range output {
		if f.Dir == "." { continue }
		if !source_dirs[f.Dir] {
			del[f.Dir] = true
		}
	}

	return mod, del
}

func support_files(root string, age time.Time) []string {
	var list []string

	if !path_exists(root) {
		return list
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		name   := info.Name()
		prefix := name[0:1]

		if prefix == "." || prefix == "_" || info.IsDir() {
			return nil
		}

		if info.ModTime().After(age) {
			name = name[0:len(name) - len(filepath.Ext(name))]
			list = append(list, name)
		}

		return nil
	})

	return list
}