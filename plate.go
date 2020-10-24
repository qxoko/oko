package main

import (
	"strings"
	"encoding/json"
)

var PlateList = make(map[string]*Plate)

type Plate struct {
	Extends       string

	SnippetBefore []string `json:"snippet_before"`
	SnippetAfter  []string `json:"snippet_after"`

	BodyBefore    []string `json:"body_before"`
	BodyAfter     []string `json:"body_after"`

	Script        []string
	Style         []string

	ScriptRender  string
	StyleRender   string

	Tokens map[string]string
}

var default_plate = &Plate {
	Tokens: map[string]string {
		"h1":        `<h1 id='%s'>%s</h1>`,
		"h2":        `<h2 id='%s'>%s</h2>`,
		"h3":        `<h3 id='%s'>%s</h3>`,
		"h4":        `<h4 id='%s'>%s</h4>`,
		"h5":        `<h5 id='%s'>%s</h5>`,
		"h6":        `<h6 id='%s'>%s</h6>`,
		"image":     `<img src='%s'>`,
		"quote":     `<blockquote>%s</blockquote>`,
		"divider":   `<hr>`,
		"paragraph": `<p>%s</p>`,
		"ul":        `<ul>%s</ul>`,
		"list":      `<li>%s</li>`,
		"code":      `<pre><code %s>%s</code></pre>`,
	},
}

func plate_entry(p *Plate, v string) string {
	if value, ok := p.Tokens[v]; ok {
		return value
	}
	if value, ok := default_plate.Tokens[v]; ok {
		return value
	}
	return ""
}

func plate_entry_offset(p *Plate, tok *Token) string {
	id, def := convert_token_offset(tok)

	if value, ok := p.Tokens[id]; ok {
		return value
	}
	if tok.Type == HEADING {
		if value, ok := default_plate.Tokens[id]; ok {
			return value
		}
	}
	if value, ok := default_plate.Tokens[def]; ok {
		return value
	}

	return ""
}

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
		// @error
		panic(sub_sprint(`failed to parse JSON in "%s"\nerror: "%s"`, path, err.Error()))
	}

	// do this in case the child plate has no
	// need for tokens - the map doesn't get
	// initialised and the combining step fails
	// with an unhelpful error
	if plate.Tokens == nil {
		plate.Tokens = make(map[string]string)
	}

	if plate.Extends != "" {
		var extend Plate

		err  := json.Unmarshal(load_file_bytes(plate_path(plate.Extends)), &extend)

		if err != nil {
			// @error
			panic(sub_sprint(`failed to parse JSON in "%s"\nerror: "%s"`, path, err.Error()))
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
	return render_stackable(list, def, `<link rel='stylesheet' type='text/css' href='%s'/>`)
}

func render_script(list []string, def string) string {
	if len(list) == 0 {
		return def
	}
	return render_stackable(list, def, `<script type='text/javascript' src='%s' defer></script>`)
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