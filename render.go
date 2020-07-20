package main

import (
	"os"
	"sort"
	"bufio"
	"strings"
	"unicode"
	"path/filepath"
)

var default_plate = &Plate {
	Tokens: map[string]string {
		"h1":        `<h1 id='${v}'>${v}</h1>`,
		"h2":        `<h2 id='${v}'>${v}</h2>`,
		"h3":        `<h3 id='${v}'>${v}</h3>`,
		"h4":        `<h4 id='${v}'>${v}</h4>`,
		"h5":        `<h5 id='${v}'>${v}</h5>`,
		"h6":        `<h6 id='${v}'>${v}</h6>`,
		"image":     `<img src='${v}'>`,
		"quote":     `<blockquote>${v}</blockquote>`,
		"divider":   `<hr>`,
		"paragraph": `<p>${v}</p>`,
		"ul":        `<ul>${v}</ul>`,
		"list":      `<li>${v}</li>`,
		"code":      `<pre><code>${v}</code></pre>`,
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

func render(p *Page) {
	if plate_name, ok := p.Vars["plate"]; ok {
		p.Plate = load_plate(plate_name)
	} else {
		p.Plate = default_plate
		p.Style = config.Style
	}

	var body strings.Builder
	var body_inside strings.Builder

	if len(p.Plate.SnippetBefore) > 0 {
		for _, s := range p.Plate.SnippetBefore {
			body.WriteString(snippet(p, s))
		}
	}

	if len(p.Plate.BodyBefore) > 0 {
		for _, s := range p.Plate.BodyBefore {
			body_inside.WriteString(snippet(p, s))
		}
	}

	body_inside.WriteString(recurse_render(p, nil))

	if len(p.Plate.BodyAfter) > 0 {
		for _, s := range p.Plate.BodyAfter {
			body_inside.WriteString(snippet(p, s))
		}
	}

	if b, ok := p.Plate.Tokens["body"]; ok {
		body.WriteString(sub_content(b, body_inside.String()))
	} else {
		body.WriteString(body_inside.String())
	}

	if len(p.Plate.SnippetAfter) > 0 {
		for _, s := range p.Plate.SnippetAfter {
			body.WriteString(snippet(p, s))
		}
	}

	var favicon string
	var title   string

	if f, ok := p.Vars["favicon"]; ok {
		favicon = make_favicon(f)
	} else {
		favicon = config.Favicon
	}

	if f, ok := p.Vars["title"]; ok {
		title = f
	} else {
		title = config.Title
	}

	file, err := os.Create(p.OutputPath)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	writer.WriteString(`<!DOCTYPE html><html><head><title>`)
	writer.WriteString(title)
	writer.WriteString(`</title><meta charset='utf-8'>`)
	writer.WriteString(favicon)
	writer.WriteString(render_style(p.Style,   p.Plate.StyleRender))
	writer.WriteString(render_script(p.Script, p.Plate.ScriptRender))
	writer.WriteString(meta(p))
	writer.WriteString(`</head><body>`)
	writer.WriteString(mapmap(body.String(), p.Vars, true))
	writer.WriteString(`</body></html>`)

	writer.Flush()
}

func skip_block(the_page *Page, active_block *Token) {
	the_list := the_page.List

	for {
		tok := the_list.Next()

		if tok == nil {
			break
		}

		if tok.Type > tok_if_statements {
			skip_block(the_page, active_block)
			continue
		}

		if tok.Type == BLOCK_START {
			skip_block(the_page, active_block)
			continue
		}

		if tok.Type == BLOCK_CLOSE {
			break
		}
	}
}

func recurse_render(the_page *Page, active_block *Token) string {
	var content strings.Builder

	the_list := the_page.List
	plate    := the_page.Plate

	for {
		tok := the_list.Next()

		if tok == nil {
			break
		}

		if tok.Type > tok_if_statements {
			if check_if_statement(the_page, tok) {
				content.WriteString(recurse_render(the_page, tok))
			} else {
				skip_block(the_page, tok)
			}
			continue
		}

		if tok.Type < tok_inline_format && tok.Type > tok_headings {
			tok.Text = inlines(tok.Text)
		}

		switch tok.Type {
			case ERROR:
				render_error(the_page, tok, "parser error")

			case FUNCTION:
				content.WriteString(do_script(the_page, tok.Text))

			case LIST_ENTRY:
				var list_buffer strings.Builder

				for tok.Type == LIST_ENTRY {
					list_buffer.WriteString(sub_content(plate_entry(plate, "list"), the_list.Peek().Text))

					tok = the_list.Lookahead()

					if tok == nil || tok.Type != LIST_ENTRY {
						break
					}

					tok = the_list.Next()
				}

				p := plate_entry(plate, "ul")

				content.WriteString(sub_content(p, inlines(list_buffer.String())))

				continue

			case SNIPPET:
				if filepath.Ext(tok.Text) == "" {
					content.WriteString(snippet(the_page, tok.Text))
				} else {
					// @todo error
					content.WriteString(string(load_file_bytes(filepath.Join("_data/snippets", tok.Text))))
				}
				continue

			case IMPORT:
				n := strings.SplitN(tok.Text, " ", 2)

				var t string

				if len(n) > 1 {
					t = strings.TrimSpace(n[1])
				} else {
					t = "import"
				}

				if v, ok := plate.Tokens[t]; ok {
					if p, ok := PageList[n[0]]; ok {
						content.WriteString(inlines(mapmap(v, p.Vars, true)))
					} else {
						warning("skipped import " + n[0])
					}
				} else {
					render_error(the_page, tok, "failed to import")
				}

				continue

			case MEDIA:
				content.WriteString(media(tok.Text))
				continue

			case HTML_SNIPPET:
				content.WriteString(tok.Text)
				continue

			case BLOCK_CODE:
				tok = the_list.Next()

				text := inline_code_sub(tok.Text)

				content.WriteString(sub_content(plate_entry(plate, "code"), text))
				continue

			case BLOCK_START:
				block_plate   := plate_entry(plate, tok.Text)
				child_content := recurse_render(the_page, tok)

				if block_plate != "" {
					child_content = sub_content(block_plate, child_content)
				}

				content.WriteString(mapmap(child_content, tok.Vars, false))
				continue

			case BLOCK_CLOSE:
				return content.String()
		}

		p := plate_entry(plate, tok.Type.String())

		if tok.Type < tok_headings {
			clean_text := strip_inlines(tok.Text)
			dirty_text := inlines(tok.Text)

			content.WriteString(sub_sprint(p, id_maker(clean_text), dirty_text))
			continue
		}

		if tok.Type == IMAGE {
			content.WriteString(sub_content(p, image_checker(tok.Text)))
			continue
		}

		content.WriteString(sub_content(p, tok.Text))
	}

	return content.String()
}

// adds media_prefix_path (oko.json) to
// images unless image is external or already
// has prefix
func image_checker(v string) string {
	if !strings.HasPrefix(v, `http`) && !strings.HasPrefix(v, config.ImagePrefix) {
		v = config.ImagePrefix + v
	}
	return v
}

func id_maker(source string) string {
	var new strings.Builder

	for _, c := range source {
		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			new.WriteRune(c)
			continue
		}
		if unicode.IsSpace(c) || c == '-' {
			new.WriteRune('-')
			continue
		}
	}

	return strings.ToLower(new.String())
}

func mapmap(source string, ref_map map[string]string, hard bool) string {
	if strings.IndexRune(source, '$') < 0 {
		return source
	}

	input := source
	list  := make(map[string]string)

	for {
		pos := strings.IndexRune(input, '$')

		if pos < 0 {
			break
		}

		if input[pos+1] == '{' {
			end     := strings.IndexRune(input[pos+1:], '}')
			end_pos := pos + end + 2

			if end > 0 {
				v := input[pos:end_pos]
				list[v] = v
			} else {
				panic("bad variable") // @error
			}

			input = input[end_pos:]
		} else {
			input = input[pos+1:]
		}
	}

	for _, variable := range list {
		id := variable[2:len(variable)-1]

		if value, ok := ref_map[id]; ok {
			if strings.Contains(id, `image`) { // do image things in variables
				value = image_checker(value)
			}
			source = strings.ReplaceAll(source, variable, value)
		} else if hard {
			source = strings.ReplaceAll(source, variable, "")
		}
	}

	return source
}

func sub(source, r, v string) string {
	return strings.ReplaceAll(source, `${` + r + `}`, v)
}

func sub_content(source, v string) string {
	return strings.ReplaceAll(source, `${v}`, v)
}

func sub_sprint(source string, v ...string) string {
	for _, x := range v {
		source = strings.Replace(source, `${v}`, x, 1)
	}
	return source
}



var SnippetText = make(map[string]string)
var SnippetList = make(map[string]*Page)

func render_snippet(p *Page) string {
	var body strings.Builder
	var body_inside strings.Builder

	if len(p.Plate.SnippetBefore) > 0 {
		for _, s := range p.Plate.SnippetBefore {
			body.WriteString(snippet(p, s))
		}
	}

	if len(p.Plate.BodyBefore) > 0 {
		for _, s := range p.Plate.BodyBefore {
			body_inside.WriteString(snippet(p, s))
		}
	}

	body_inside.WriteString(recurse_render(p, nil))

	if len(p.Plate.BodyAfter) > 0 {
		for _, s := range p.Plate.BodyAfter {
			body_inside.WriteString(snippet(p, s))
		}
	}

	if b, ok := p.Plate.Tokens["body"]; ok {
		body.WriteString(sub_content(b, body_inside.String()))
	} else {
		body.WriteString(body_inside.String())
	}

	if len(p.Plate.SnippetAfter) > 0 {
		for _, s := range p.Plate.SnippetAfter {
			body.WriteString(snippet(p, s))
		}
	}

	return mapmap(body.String(), p.Vars, false)
}

func snippet(parent *Page, name string) string {
	if body, ok := SnippetText[name]; ok {
		return body
	}
	if saved_page, ok := SnippetList[name]; ok {
		saved_page.List.Reset()
		saved_page.CurrentParent = parent
		return render_snippet(saved_page)
	}

	path := filepath.Join("_data/snippets", name + ".ø")
	the_page := &Page{}

	the_page.Vars = make(map[string]string)
	the_page.CurrentParent = parent

	if !file_exists(path) {
		warning("snippet " + name + " does not exist")
		return ""
	}

	the_page.List = parser(the_page, load_file_bytes(path))

	if the_page.IsDraft {
		warning("cannot have draft snippet " + the_page.ID)
	}

	if plate_name, ok := the_page.Vars["plate"]; ok {
		the_page.Plate = load_plate(plate_name)
	} else {
		the_page.Plate = default_plate
	}

	b := render_snippet(the_page)

	if the_page.List.IsCommittable {
		SnippetText[name] = b
	} else {
		SnippetList[name] = the_page
	}

	return b
}

func check_slash(s string) string {
	if s[len(s)-1:] != "/" {
		s += "/"
	}
	return s
}

var meta_source  = `<meta property='${v}' content='${v}'>`
var meta_descrip = `<meta name='description' content='${v}'>`

func meta(the_page *Page) string {
	var meta_block strings.Builder

	if _, ok := the_page.Meta["title"]; !ok {
		the_page.Meta["title"] = the_page.Vars["title"]
	}
	if _, ok := the_page.Meta["description"]; !ok {
		if v, ok := config.Meta["description"]; ok {
			the_page.Meta["description"] = v
		}
	}
	if _, ok := the_page.Meta["image"]; !ok {
		if v, ok := config.Meta["image"]; ok {
			the_page.Meta["image"] = v
		}
	}

	// domain
	canon_path := config.Domain + the_page.URLPath
	meta_block.WriteString(sub_content(`<link rel='canonical' href='${v}'>`, canon_path))
	meta_block.WriteString(sub_sprint(meta_source, "og:url", canon_path))

	domain := check_slash(config.Domain)

	// generic opengraph entries
	for tag, value := range the_page.Meta {
		switch tag {
			case "image":
				if strings.HasPrefix(value, "/") {
					value = config.Domain + value
				} else {
					value = domain + value
				}

			case "description":
				meta_block.WriteString(sub(meta_descrip, "v", value))
		}

		meta_block.WriteString(sub_sprint(meta_source, "og:" + tag, value))
	}

	// twitter
	needs_media_card := false

	var twitter_creator string
	var twitter_site    string

	if value, ok := the_page.Vars["twitter_creator"]; ok {
		twitter_creator = value
	} else if value, ok := config.Meta["twitter_creator"]; ok {
		twitter_creator = value
	}

	if value, ok := the_page.Vars["twitter_site"]; ok {
		twitter_site = value
	} else if value, ok := config.Meta["twitter_site"]; ok {
		twitter_site = value
	}

	if twitter_creator != "" {
		meta_block.WriteString(sub_sprint(meta_source, "twitter:creator", twitter_creator))
		needs_media_card = true
	}

	if twitter_site != "" {
		meta_block.WriteString(sub_sprint(meta_source, "twitter:site", twitter_site))
		needs_media_card = true
	}

	if needs_media_card {
		meta_block.WriteString(`<meta property='twitter:card' content='summary_large_image'>`)
	}

	return meta_block.String()
}

func sitemap(path string) {
	ordered    := make([]string, len(PageList))
	url_source := `<url><loc>${v}${v}</loc></url>`

	for _, page := range PageList {
		ordered = append(ordered, sub_sprint(url_source, config.Domain, page.URLPath))
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i] < ordered[j]
	})

	file, err := os.Create(path)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	writer.WriteString(`<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">`)

	for _, page := range ordered {
		writer.WriteString(page)
	}

	writer.WriteString(`</urlset>`)

	writer.Flush()
}

func make_favicon(f string) string {
	var tag string

	switch filepath.Ext(f) {
		case ".ico": tag = `<link rel='icon' type='image/x-icon' href='${v}'>`
		case ".png": tag = `<link rel='icon' type='image/png' href='${v}'>`
		case ".gif": tag = `<link rel='icon' type='image/gif' href='${v}'>`
		default: panic("bad favicon format")
	}

	return sub_content(tag, f)
}