package main

import (
	"os"
	"sort"
	"bufio"
	"strings"
	"path/filepath"
)

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

			case FUNCTION:
				content.WriteString(tok.Text)

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

			content.WriteString(sub_sprint(p, make_element_id(clean_text), dirty_text))
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


//
// Snippets
//
var SnippetText = make(map[string]string)
var SnippetList = make(map[string]*Page)

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

//
// Meta
//
func check_slash(s string) string {
	if s[len(s)-1:] != "/" {
		s += "/"
	}
	return s
}

var meta_source  = `<meta property='%s' content='%s'>`
var meta_descrip = `<meta name='description' content='%s'>`

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
	meta_block.WriteString(sub_content(`<link rel='canonical' href='%s'>`, canon_path))
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
	url_source := `<url><loc>%s%s</loc></url>`

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