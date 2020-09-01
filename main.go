package main

import (
	"os"
	"strings"
)

var config   *Config
var AllFiles map[string]*File

func oko() {
	input,           _ := walk(".")
	output, output_age := walk(config.Output)

	AllFiles = input

	// parse all files
	for _, file := range input {
		if file.Type == MARKUP {
			create_page(file)
		}
	}

	// execute all javascript
	for _, file := range input {
		if file.Type == MARKUP && file.Page.HasFunction {
			do_functions(file.Page)
		}
	}

	// set flags on file list
	set_update_flags(input, output, support_files(output_age))

	if config.DoAllPages {
		for _, file := range input {
			if file.Type == MARKUP {
				file.Action = NEEDS_UPDATE
			}
		}
	}

	// create directories
	for _, file := range input {
		if file.Type == DIR && file.Action == NEEDS_UPDATE {
			mkdir(file.OutputPath)
		}
	}

	// create/delete files
	for _, file := range input {
		if file.Type == DIR {
			continue
		}

		switch file.Action {
			case NEEDS_UPDATE:
				if file.Type == MARKUP {
					print(sub_sprint("render file %s\n", file.SourcePath))
					render(file)
				} else {
					print(sub_sprint("update file %s\n", file.SourcePath))
					copy_file(file.SourcePath, file.OutputPath)
				}

			case NEEDS_DELETE:
				delete_file(file.OutputPath)
		}
	}

	// delete directories
	for _, file := range input {
		if file.Type == DIR && file.Action == NEEDS_DELETE {
			delete_file(file.OutputPath)
		}
	}

	print_warnings()
}

func set_update_flags(input, output map[string]*File, support map[string]bool) {
	// set updates by modification time
	for id, i := range input {
		if o, ok := output[id]; ok {
			if i.Mod.After(o.Mod) {
				i.Action = NEEDS_UPDATE
			}
		} else {
			i.Action = NEEDS_UPDATE
		}
	}

	for id, o := range output {
		if i, ok := input[id]; ok {
			if i.IsDraft && !config.ShowDrafts {
				i.Action = NEEDS_DELETE
			}
		} else {
			o.Action = NEEDS_DELETE
			input[id] = o
		}
	}

	// dependencies / drafts
	for _, i := range input {
		if i.Type == MARKUP {
			if i.Page.IsDraft && !config.ShowDrafts { // exclude drafts
				i.Action = NONE
				continue
			}
		}

		if i.Type == DIR {
			any_remain := false

			print(i.SourcePath, "\n")

			for _, f := range i.Children {
				print("    ", f.SourcePath, "\n")

				if f.Action != NEEDS_DELETE {
					any_remain = true
					break
				}
			}

			if !any_remain {
				i.Action = NEEDS_DELETE
			}
		}
	}

	// deps
	for s, b := range support {
		if !b { continue }
		if list, ok := ExternalDeps[s]; ok {
			for _, f := range list {
				input[f].Action = NEEDS_UPDATE
			}
		}
	}

	for _, i := range input {
		if i.Type == MARKUP && i.Action == NEEDS_UPDATE {
			for id, _ := range i.Page.Deps {
				input[id].Action = NEEDS_UPDATE
			}
		}
	}
}

func main() {
	new_config   := false
	do_all_pages := false
	show_drafts  := false
	watch_files  := false
	run_server   := false

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, `--`) {
			switch arg[2:] {
				case "new-config": new_config   = true
				case "all":        do_all_pages = true
				case "drafts":     show_drafts  = true

				case "watch":
					watch_files = true

				case "serve":
					watch_files = true
					run_server  = true
			}
		} else if strings.HasPrefix(arg, `-`) {
			for _, c := range arg[1:] {
				switch c {
					case 'a':
						do_all_pages = true

					case 'd':
						show_drafts  = true

					case 's':
						watch_files  = true
						run_server   = true

					case 'w':
						watch_files  = true

					case 'D':
						do_all_pages = true
						show_drafts  = true
						watch_files  = true
						run_server   = true
				}
			}
		} else {
			panic("bad argument") // @error
		}
	}

	if new_config {
		make_new_config_file()
		print("[ø] created project file!")
		return
	}

	config = load_config()

	if config == nil {
		print("[ø] not an oko project!")
		return
	}

	if do_all_pages {
		config.DoAllPages = true
	}
	if show_drafts {
		config.ShowDrafts = true
	}

	oko()

	if run_server || watch_files {
		// disable do_all_pages in here
	}
}