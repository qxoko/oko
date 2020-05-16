package main

import (
	"log"
	"strings"
	"encoding/json"
)

type Config struct {
	Domain  string
	Output  string
	Favicon string
	Title   string

	StyleRender string

	DoAllPages bool `json:"do_all_pages"`
	Sitemap bool

	Style   []string
	Include []string

	Meta map[string]string
}

func load_config() *Config {
	var config Config

	p := "_data/oko.json"

	if !file_exists(p) {
		return nil
	}

	err := json.Unmarshal(load_file_bytes(p), &config)

	if err != nil {
		panic(err)
	}

	if config.Domain == "" {
		panic("no domain name in _data/oko.json!")
	}

	if config.Output == "" {
		config.Output = "public" // set a reasonable default
	}

	if config.Favicon != "" {
		config.Favicon = make_favicon(config.Favicon)
	}

	config.StyleRender = render_style(config.Style, ``)

	return &config
}

type Plate struct {
	Extends       string

	SnippetBefore []string `json:"snippet_before"`
	SnippetAfter  []string `json:"snippet_after"`

	Script        []string
	Style         []string

	ScriptRender  string
	StyleRender   string

	Tokens map[string]string
}

var PlateList = make(map[string]*Plate)

func plate_path(name string) string {
	return "_data/plates/" + name + ".json"
}

func load_plate(name string) *Plate {
	if plate, ok := PlateList[name]; ok {
		return plate
	}

	var plate Plate

	path := plate_path(name)
	err  := json.Unmarshal(load_file_bytes(path), &plate)

	if err != nil {
		log.Fatalf("failed to parse JSON in %q\nerror: %q", path, err)
	}

	if plate.Extends != "" {
		var extend Plate

		err  := json.Unmarshal(load_file_bytes(plate_path(plate.Extends)), &extend)

		if err != nil {
			log.Fatalf("failed to parse JSON in %q\nerror: %q", path, err)
		}

		// merge
		if len(plate.SnippetBefore) == 0 {
			plate.SnippetBefore = extend.SnippetBefore
		}
		if len(plate.SnippetAfter) == 0 {
			plate.SnippetAfter = extend.SnippetAfter
		}
		if len(plate.Script) == 0 {
			plate.Script = extend.Script
		}
		if len(plate.Style) == 0 {
			plate.Style = extend.Style
		}

		for key, val := range extend.Tokens {
			if v, ok := plate.Tokens[key]; !ok {
				plate.Tokens[key] = val
			} else if v == "" {
				delete(plate.Tokens, key)
			}
		}
	}

	plate.StyleRender  = render_style(plate.Style,   config.StyleRender)
	plate.ScriptRender = render_script(plate.Script, ``)

	PlateList[name] = &plate

	return &plate
}

func render_style(list []string, def string) string {
	if len(list) == 0 {
		return def
	}
	return render_stackable(list, def, `<link rel='stylesheet' type='text/css' href='${v}'/>`)
}

func render_script(list []string, def string) string {
	if len(list) == 0 {
		return def
	}
	return render_stackable(list, def, `<script type='text/javascript' src='${v}' defer></script>`)
}

func render_stackable(list []string, def, source string) string {
	var r strings.Builder

	for _, item := range list {
		if item == "default" {
			r.WriteString(def)
		} else {
			r.WriteString(sub_content(source, item))
		}
	}

	return r.String()
}