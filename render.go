package main

import (
	"fmt"
	"strings"
	"path/filepath"
)

var default_plate = &Plate {
	Tokens: map[string]string {
		"h1":        `<h1>${v}</h1>`,
		"h2":        `<h2>${v}</h2>`,
		"h3":        `<h3>${v}</h3>`,
		"image":     `<img src="${v}">`,
		"quote":     `<QUOTE src="${v}">`,
		"divider":   `<hr>`,
		"paragraph": `<p>${v}</p>`,
		"ul":        `<ul>${v}</ul>`,
		"list":      `<li>${v}</li>`,
		"code":      `<pre><code>${v}</code></pre>`,
	},
}

var page_source = `<!DOCTYPE html><html><head><title>${title}</title><meta charset="utf-8">${favicon}${_style}${_meta}</head><body>${body}${_script}</body></html>`

func plate_entry(p *Plate, v string) string {
	if value, ok := p.Tokens[v]; ok {
		return value
	}
	if value, ok := default_plate.Tokens[v]; ok {
		return value
	}
	fmt.Println("bad token in plate", v)
	return ""
}

func render(p *Page) {
	if p.IsRendered {
		return
	}

	if plate_name, ok := p.Vars["plate"]; ok {
		p.Plate = load_plate(plate_name)
	} else {
		p.Plate = default_plate
		p.Style = config.Style
	}

	var body strings.Builder

	if len(p.Plate.SnippetBefore) > 0 {
		for _, s := range p.Plate.SnippetBefore {
			body.WriteString(snippet(s))
		}
	}

	inside := recurse_render(p, nil)

	if b, ok := p.Plate.Tokens["body"]; ok {
		body.WriteString(sub_content(b, inside))
	} else {
		body.WriteString(inside)
	}

	if len(p.Plate.SnippetAfter) > 0 {
		for _, s := range p.Plate.SnippetAfter {
			body.WriteString(snippet(s))
		}
	}

	if _, ok := p.Vars["favicon"]; !ok {
		p.Vars["favicon"] = config.Favicon
	}

	p.Vars["body"]    = mapmap(body.String(), p.Vars, true)
	p.Vars["_style"]  = render_style(p.Style,   p.Plate.StyleRender)
	p.Vars["_script"] = render_script(p.Script, p.Plate.ScriptRender)
	p.Vars["_meta"]   = meta(p)

	p.Vars["page_path"] = filepath.ToSlash("/" + p.ID)

	content := mapmap(page_source, p.Vars, true)

	p.Vars["full_render"] = content

	write_file(p.OutputPath, content)

	p.IsRendered = true
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
			continue // @todo if statements
		}

		if tok.Type < tok_inline_format && tok.Type > tok_headings {
			tok.Text = inlines(tok.Text)
		}

		switch tok.Type {
			case ERROR:
				render_error(the_page, tok, "parser error")

			case FUNCTION:
				render_error(the_page, tok, "functions unsupported")

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
				content.WriteString(snippet(tok.Text))
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
						content.WriteString(mapmap(v, p.Vars, false))
					} else {
						render_error(the_page, tok, "failed to import")
					}
				}

				continue

			case VIDEO:
				content.WriteString(video(tok.Text))
				continue

			case HTML_SNIPPET:
				content.WriteString(tok.Text)
				continue

			case BLOCK_CODE:
				tok = the_list.Next()
				content.WriteString(sub_content(plate_entry(plate, "code"), tok.Text))
				continue

			case BLOCK_START:
				block_plate   := plate_entry(plate, tok.Text)
				child_content := sub_content(block_plate, recurse_render(the_page, tok))
				content.WriteString(mapmap(child_content, tok.Vars, false))
				continue

			case BLOCK_CLOSE:
				return content.String()
		}

		p := plate_entry(plate, tok.Type.String())

		if tok.Type < tok_headings {
			// id := strings.ReplaceAll(strings.ToLower(strip_inlines(tok.Text)), " \t", "-")
			// content.WriteString(sub_sprint(p, id, inlines(tok.Text)))

			// @todo reinstate ids

			content.WriteString(sub_content(p, inlines(tok.Text)))
			continue
		}

		content.WriteString(sub_content(p, tok.Text))
	}

	return content.String()
}

// @todo rewrite this for _speeeeed_
func mapmap(source string, ref_map map[string]string, hard bool) string {
	if strings.IndexRune(source, '$') < 0 {
		return source
	}

	input := source
	list  := make(map[string]string)

	for {
		pos := strings.IndexRune(input, '$')

		if !(pos >= 0) {
			break
		}

		if input[pos+1] == 123 { // "{"
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



var SnippetList = make(map[string]string)

func snippet(name string) string {
	if v, ok := SnippetList[name]; ok {
		return v
	}

	path := filepath.Join("_data/snippets/", name + ".ø")
	page := &Page{}

	page.Vars = make(map[string]string)

	if list, is_draft := parser(page, load_file_bytes(path)); !is_draft {
		page.List = list
	} else {
		return ""
	}

	if plate_name, ok := page.Vars["plate"]; ok {
		page.Plate = load_plate(plate_name)
	} else {
		page.Plate = default_plate
	}

	var body strings.Builder

	if len(page.Plate.SnippetBefore) > 0 {
		for _, s := range page.Plate.SnippetBefore {
			body.WriteString(snippet(s))
		}
	}

	inside := inlines(recurse_render(page, nil))

	if b, ok := page.Plate.Tokens["body"]; ok {
		body.WriteString(sub_content(b, inside))
	} else {
		body.WriteString(inside)
	}

	if len(page.Plate.SnippetAfter) > 0 {
		for _, s := range page.Plate.SnippetAfter {
			body.WriteString(snippet(s))
		}
	}

	b := mapmap(body.String(), page.Vars, false)

	SnippetList[name] = b

	return b
}

func check_slash(s string) string {
	if s[len(s)-1:] != "/" {
		s += "/"
	}
	return s
}

var meta_source  = `<meta property="${v}" content="${v}">`
var meta_descrip = `<meta name="description" content="${v}">`

func meta(the_page *Page) string {
	var meta_block strings.Builder

	if _, ok := the_page.Meta["title"]; !ok {
		the_page.Meta["title"] = the_page.Vars["title"]
	}
	if _, ok := the_page.Meta["description"]; !ok {
		the_page.Meta["description"] = config.Meta["description"]
	}
	if _, ok := the_page.Meta["image"]; !ok {
		the_page.Meta["image"] = config.Meta["image"]
	}

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

	// domain
	meta_block.WriteString(sub_sprint(meta_source, "og:url", domain + the_page.URLPath))

	// twitter
	if c, ok := config.Meta["twitter"]; ok {
		meta_block.WriteString(sub_sprint(meta_source, "twitter:creator", c))
		meta_block.WriteString(`<meta property="twitter:card" content="summary_large_image">`)
	}

	return meta_block.String()
}

func sitemap() {
	sitemap_source := `<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">${v}</urlset>`
	url_source := `<url><loc>${v}${v}</loc></url>`

	var url_block strings.Builder

	for _, page := range PageList {
		url_block.WriteString(sub_sprint(url_source, config.Domain, page.URLPath))
	}

	final := sub_content(sitemap_source, url_block.String())

	write_file(filepath.Join(config.Output, "sitemap.xml"), final)
}