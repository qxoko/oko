package main

import (
	"os"
	"fmt"
	"flag"
	"sort"
	"time"
	"path/filepath"
)

var config *Config = load_config()

func do_pages(do_all_pages bool) {
	source, _   := walk(".", ".ø", ".html")
	output, age := walk(config.Output, ".html")

	file_mod, file_del := compare_files(source, output)
	path_mod, path_del := compare_dirs(source, output, file_mod, file_del)

	snippets := support_files("_data/snippets", age)
	plates   := support_files("_data/plates",   age)

	sum := len(file_mod) + len(plates) + len(snippets)

	if sum > 0 || do_all_pages {
		for _, file := range source {
			if file.Format != OKO {
				continue
			}

			the_page := make_page(file)
			bytes    := load_file_bytes(the_page.SourcePath)

			if list, is_draft := parser(the_page, bytes); is_draft {
				file.IsDraft = true
			} else {
				the_page.List = list
			}
		}

		for path, _ := range path_mod {
			mkdir(filepath.Join(config.Output, path))
		}

		if do_all_pages {
			file_mod  = make([]*File_Info, 0, len(source))
			for _, f := range source {
				file_mod = append(file_mod, f)
			}
		} else {
			if len(snippets) > 0 {
				for _, s := range snippets {
					for _, e := range DepTree["snip_" + s] {
						file_mod = append(file_mod, source[e])
					}
				}
			}
			if len(plates) > 0 {
				for _, p := range plates {
					for _, e := range DepTree["plate_" + p] {
						file_mod = append(file_mod, source[e])
					}
				}
			}
		}

		for _, file := range file_mod {
			if file.Format == HTML {
				copy_file(file.Path, filepath.Join(config.Output, file.Path))
				continue
			}

			if file.IsDraft {
				continue
			}

			render(PageList[file.ID])
		}
	}

	if len(file_del) > 0 || len(path_del) > 0 {
		for _, file := range file_del {
			delete_file(filepath.Join(config.Output, file.Path))
		}
		for path, _ := range path_del {
			delete_file(filepath.Join(config.Output, path))
		}
	}

	// report results
	if len(file_mod) > 0 {
		if config.Sitemap {
			sitemap()
		}

		sort.SliceStable(file_mod, func(i, j int) bool {
			return file_mod[i].ID < file_mod[j].ID
		})

		fmt.Println("[ø] updated pages\n")

		for _, file := range file_mod {
			fmt.Println("   ", file.ID)
		}

		fmt.Println()
	}
}

func support_files(root string, age time.Time) []string {
	var list []string

	if !path_exists(root) {
		return list
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		name   := info.Name()
		prefix := name[0:1]

		if prefix == "." || prefix == "_" {
			return nil
		}

		if !info.IsDir() && info.ModTime().After(age) {
			name = name[0:len(name) - len(filepath.Ext(name))]
			list = append(list, name)
		}

		return nil
	})

	return list
}

func do_static_files() {
	var list []string

	for _, file := range config.Include {

		// directories
		if path_exists(file) {
			source, _ := walk(file)
			output, _ := walk(filepath.Join(config.Output, file))

			for _, f := range source {
				f.Dir = filepath.Join(file, f.Dir) // @hack
			}
			for _, f := range output {
				f.Dir = filepath.Join(file, f.Dir) // @hack
			}

			file_mod, file_del := compare_files(source, output)
			path_mod, path_del := compare_dirs(source, output, file_mod, file_del)

			for path, _ := range path_mod {
				mkdir(filepath.Join(config.Output, path))
			}

			for _, n := range file_mod {
				s := filepath.Join(file, n.Path)
				o := filepath.Join(config.Output, s)
				copy_file(s, o)
				list = append(list, s)
			}

			for _, n := range file_del {
				delete_file(filepath.Join(config.Output, file, n.Path))
			}

			for path, _ := range path_del {
				mkdir(filepath.Join(config.Output, path))
			}
		} else {
			// single files
			out_path := filepath.Join(config.Output, file)

			if info, ok := file_data(file); ok {
				do_copy := false

				if out, ok := file_data(out_path); ok {
					if info.ModTime().After(out.ModTime()) {
						do_copy = true
					}
				} else {
					do_copy = true
				}

				if do_copy {
					copy_file(file, out_path)
					list = append(list, file)
				}

			} else {
				out_path := filepath.Join(config.Output, file)

				if file_exists(out_path) {
					delete_file(out_path)
					continue
				}

				warning("no file to include: " + file)
			}
		}
	}

	if len(list) > 0 {
		fmt.Println("[ø] updated static files\n")

		for _, file := range list {
			fmt.Println("   ", file)
		}

		fmt.Println()
	}
}

func main() {
	FLAG_ALL := flag.Bool("all", false, "")

	flag.Parse()

	if *FLAG_ALL {
		do_pages(true)
	} else {
		do_pages(false)
	}

	do_static_files()

	print_warnings()
}