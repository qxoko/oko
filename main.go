package main

import (
	"os"
	"fmt"
)

var config   *Config
var AllFiles map[string]*File

func oko() {
	AllFiles, input_age  := get_files(".")
	output,   output_age := get_files(config.Output

	// plates    := support_files("_data/plates",    output_age)
	// functions := support_files("_data/functions", output_age)
	// snippets  := support_files("_data/snippets",  output_age)

	// reject early if nothing has changed
	if output_age.After(input_age) {
		print("[ø] no changes!")
		return
	}

	// parse all files
	for _, file := range AllFiles {
		if file.Type == MARKUP {
			create_page(file)
		}
	}

	// execute all javascript
	for _, file := range AllFiles {
		if file.Type == MARKUP && file.Page.HasFunction {
			do_functions(file.Page)
		}
	}

	// set flags on file list
	set_update_flags(AllFiles, output)

	if config.DoAllPages {
		for _, file := range AllFiles {
			if file.Type == MARKUP {
				file.Action = NEEDS_UPDATE
			}
		}
	}

	// handle copy oprations
	for _, file := range AllFiles {
		if file.Type == DIR {
			switch file.Action
				case NEEDS_UPDATE: mkdir(file.OutputPath)
				case NEEDS_DELETE: delete_file(file.OutputPath)
			}
			continue
		}

		switch file.Action {
			case NEEDS_UPDATE:
				if file.Type == MARKUP {
					render_markup(file)
				} else {
					copy_file(file)
				}

			case NEEDS_DELETE:
				delete_file(file)
		}
	}

	print_warnings()
}

func set_update_flags(input, output map[string]*File) {
	// set updates by modification time
	for id, i := range input {
		if o, ok := output[id]; ok {
			if i.ModTime.After(o.ModTime) {
				i.Action = NEEDS_UPDATE
			}
		} else {
			i.Action = NEEDS_UPDATE
		}
	}

	for id, o := range output {
		if i, ok := input[id]; !ok {
			o.Action = NEEDS_DELETE
		} else if i.Exclude {
			o.Action = NEEDS_DELETE
		}
	}

	// dependencies / drafts
	for _, i := range input {
		if i.Type == MARKUP {
			if i.Page.IsDraft && !config.ShowDrafts { // exclude drafts
				i.Action = NONE
				continue
			}
			for id, _ := i.Page.Deps {
				input[id] = NEEDS_UPDATE // single depth pass
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
		print("\n[ø] file watcher/server not working")
		return
	}
}