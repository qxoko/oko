package main

import (
	"os"
	"fmt"
	"sort"
	"path/filepath"
)

func do_pages() {
	source, _   := walk(".", ".ø", ".html")
	output, age := walk(config.Output, ".html")

	file_mod, file_del := compare_files(source, output)
	path_mod, path_del := compare_dirs(source, output, file_mod, file_del)

	snippets := support_files("_data/snippets", age)
	plates   := support_files("_data/plates",   age)

	do_parse := (len(file_mod) + len(snippets) + len(plates)) > 0 || config.DoAllPages

	if do_parse {
		for _, file := range source {
			if file.Format != OKO {
				continue
			}

			the_page := make_page(file)
			bytes    := load_file_bytes(the_page.SourcePath)

			the_page.List = parser(the_page, bytes)
		}
	}

	if config.DoAllPages {
		file_mod  = make(map[string]*File_Info, len(source))

		for _, f := range source {
			file_mod[f.ID] = f
		}

	} else {
		for _, file := range file_mod {
			if list, ok := DepTree[file.ID]; ok {
				for _, id := range list {
					file_mod[id] = source[id]
				}
			}
		}
		for _, s := range snippets {
			for _, id := range DepTree["snip_" + s] {
				file_mod[id] = source[id]
			}
		}
		for _, p := range plates {
			for _, id := range DepTree["plate_" + p] {
				file_mod[id] = source[id]
			}
		}
	}

	for path, _ := range path_mod {
		mkdir(filepath.Join(config.Output, path))
	}

	if config.Sitemap && len(file_mod) > 0 {
		sitemap()
	}

	report_list := make([]string, 0, len(file_mod))

	for ID, file := range file_mod {
		if file.Format == HTML {
			copy_file(file.Path, filepath.Join(config.Output, file.Path))
			continue
		}

		page := PageList[ID]

		if page.IsDraft {
			continue
		}

		report_list = append(report_list, ID)

		render(page)
	}

	for _, file := range file_del {
		delete_file(filepath.Join(config.Output, file.Path))
	}

	for path, _ := range path_del {
		delete_file(filepath.Join(config.Output, path))
	}

	if len(report_list) > 0 {
		sort.SliceStable(report_list, func(i, j int) bool {
			return report_list[i] < report_list[j]
		})

		fmt.Println("[ø] updated pages\n")

		for _, file := range report_list {
			fmt.Println("   ", file)
		}

		fmt.Println()
	}
}

func do_static_files() {
	if len(config.Include) == 0 {
		return
	}

	report_list := make([]string, 0, 16)

	for _, file := range config.Include {
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
				report_list = append(report_list, s)
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
					report_list = append(report_list, file)
				}

			} else {
				if file_exists(out_path) {
					delete_file(out_path)
					continue
				}

				warning("no file to include: " + file)
			}
		}
	}

	if len(report_list) > 0 {
		fmt.Println("[ø] updated static files\n")

		sort.SliceStable(report_list, func(i, j int) bool {
			return report_list[i] < report_list[j]
		})

		for _, file := range report_list {
			fmt.Println("   ", file)
		}

		fmt.Println()
	}
}

var config *Config

func main() {
	config = load_config()

	if config == nil {
		fmt.Println("[ø] not an oko project!")
		return
	}

	for _, arg := range os.Args[1:] {
		switch arg[1:] {
			case "all": config.DoAllPages = true
		}
	}

	do_pages()
	do_static_files()

	print_warnings()
}