package main

import (
	"fmt"
	"bytes"
	"strings"
	"path/filepath"
)

var default_plate = &Plate {
	Tokens: map[string]string {
		"h1":        `<h1 id="${id}">${content}</h1>`,
		"h2":        `<h2 id="${id}">${content}</h2>`,
		"h3":        `<h3 id="${id}">${content}</h3>`,
		"image":     `<img src="${content}">`,
		"quote":     `<QUOTE src="${content}">`,
		"divider":   `<hr>`,
		"paragraph": `<p>${content}</p>`,
		"ul":        `<ul>${content}</ul>`,
		"list":      `<li>${content}</li>`,
		"code":      `<pre><code>${content}</code></pre>`,
	},
}

var page_source = []byte(`<!DOCTYPE html><html><head><title>${title}</title><meta charset="utf-8">${favicon}${_style}${_meta}</head><body>${body}${_script}</body></html>`)

func plate_entry(p *Plate, v string) []byte {
	if value, ok := p.Tokens[v]; ok {
		return []byte(value)
	}
	if value, ok := default_plate.Tokens[v]; ok {
		return []byte(value)
	}
	fmt.Println("bad token in plate", v)
	return nil
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

	var body bytes.Buffer

	if len(p.Plate.SnippetBefore) > 0 {
		for _, s := range p.Plate.SnippetBefore {
			body.WriteString(snippet(s))
		}
	}

	// inside := inlines(recurse_render(p, nil))
	inside := recurse_render(p, nil)

	if b, ok := p.Plate.Tokens["body"]; ok {
		body.WriteString(string(sub_content([]byte(b), inside)))
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

	p.Vars["body"]    = string(mapmap(body.Bytes(), p.Vars, true))
	p.Vars["_style"]  = render_style(p.Style,   p.Plate.StyleRender)
	p.Vars["_script"] = render_script(p.Script, p.Plate.ScriptRender)
	p.Vars["_meta"]   = meta(p)

	p.Vars["page_path"] = filepath.ToSlash("/" + p.ID)

	content := mapmap(page_source, p.Vars, true)

	p.Vars["full_render"] = content

	write_file(p.OutputPath, content)

	fmt.Println("page: ", p.SourcePath)

	p.IsRendered = true
}

func recurse_render(the_page *Page, active_block *Token) string {
	var content bytes.Buffer

	the_list := the_page.List
	plate    := the_page.Plate

	for {
		tok := the_list.next()

		if tok == nil {
			break
		}

		if tok.Type > IF_STATEMENT {
			continue // @todo if statements
		}

		switch tok.Type {
			case ERROR:
				panic(tok.Text)

			case FUNCTION:
				fmt.Printf("%s L%s: unsupported function", the_page.ID, tok.Line)
				panic("functions not supported!")

			case LIST_ENTRY:
				var list_buffer bytes.Buffer

				for tok.Type == LIST_ENTRY {
					tok = the_list.peek()

					section := sub_content(plate_entry(plate, "list"), tok.Text)
					list_buffer.WriteString(string(section))

					tok = the_list.lookahead()

					if tok == nil || tok.Type != LIST_ENTRY {
						break
					}

					tok = the_list.next()
				}

				section := sub_content(plate_entry(plate, "ul"), list_buffer.String())
				content.WriteString(string(section))

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
						content.WriteString(mapmap([]byte(v), p.Vars, false))
					} else {
						fmt.Printf("%s L%d: %s %q\n", the_page.ID, tok.Line, "no such page to import", tok.Text) // @error
					}
				}

				continue

			case HTML_SNIPPET:
				content.WriteString(tok.Text)
				continue

			case BLOCK_CODE:
				tok = the_list.next()

				code := sub_content(plate_entry(plate, "code"), tok.Text)

				content.WriteString(string(code))
				continue

			case BLOCK_START:
				block_plate   := plate_entry(plate, tok.Text)
				child_content := sub_content(block_plate, recurse_render(the_page, tok))
				content.WriteString(string(mapmap(child_content, tok.Vars, false)))
				continue

			case BLOCK_CLOSE:
				return content.String()
		}

		value := plate_entry(plate, tok.Type.String())

		if tok.Type < HEADINGS {
			id := strings.ReplaceAll(strings.ToLower(tok.Text), " \t", "-")
			x  := sub_content(sub(value, "id", id), tok.Text)

			content.WriteString(string(x))

			continue
		}

		content.WriteString(string(sub_content(value, tok.Text)))
	}

	return content.String()
}

// @todo rewrite this for _speeeeed_
func mapmap(source []byte, ref_map map[string]string, hard bool) string {
	if bytes.IndexRune(source, '$') < 0 {
		return string(source)
	}

	input := source

	list := make(map[string][]byte)

	for {
		pos := bytes.IndexRune(input, '$')

		if !(pos >= 0) {
			break
		}

		if input[pos+1] == 123 { // "{"
			end := bytes.IndexRune(input[pos+1:], '}')
			end_pos := pos + end + 2

			if end > 0 {
				v := input[pos:end_pos]
				list[string(v)] = v
			} else {
				panic("bad variable") // @error
			}

			input = input[end_pos:]
		}
	}

	for _, variable := range list {
		id := variable[2:len(variable)-1]

		if value, ok := ref_map[string(id)]; ok {
			source = bytes.ReplaceAll(source, variable, []byte(value))
		} else if hard {
			source = bytes.ReplaceAll(source, variable, []byte(""))
		}
	}

	return string(source)
}

func sub(source []byte, r, v string) []byte {
	return bytes.ReplaceAll(source, []byte(`${` + r + `}`), []byte(v))
}

func sub_content(source []byte, v string) []byte {
	return bytes.ReplaceAll(source, []byte(`${content}`), []byte(v))
}

func sub_sprint(source []byte, v ...string) []byte {
	for _, x := range v {
		source = bytes.Replace(source, []byte(`${v}`), []byte(x), 1)
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

	var body bytes.Buffer

	if len(page.Plate.SnippetBefore) > 0 {
		for _, s := range page.Plate.SnippetBefore {
			body.WriteString(snippet(s))
		}
	}

	inside := inlines(recurse_render(page, nil))

	if b, ok := page.Plate.Tokens["body"]; ok {
		body.WriteString(string(sub_content([]byte(b), inside)))
	} else {
		body.WriteString(inside)
	}

	if len(page.Plate.SnippetAfter) > 0 {
		for _, s := range page.Plate.SnippetAfter {
			body.WriteString(snippet(s))
		}
	}

	b := string(mapmap(body.Bytes(), page.Vars, false))

	SnippetList[name] = b

	return b
}

func check_slash(s string) string {
	if s[len(s)-1:] != "/" {
		s += "/"
	}
	return s
}

var meta_source  = []byte(`<meta property="${v}" content="${v}">`)
var meta_descrip = []byte(`<meta name="description" content="${v}">`)

// @todo make these things optional

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
				meta_block.WriteString(string(sub(meta_descrip, "v", value)))
		}

		meta_block.WriteString(string(sub_sprint(meta_source, "og:" + tag, value)))
	}

	// domain
	meta_block.WriteString(string(sub_sprint(meta_source, "og:url", domain + the_page.URLPath)))

	// twitter
	if c, ok := config.Meta["twitter"]; ok {
		meta_block.WriteString(string(sub_sprint(meta_source, "twitter:creator", c)))
		meta_block.WriteString(`<meta property="twitter:card" content="summary_large_image">`)
	}

	return meta_block.String()
}

func sitemap() {
	sitemap_source := []byte(`<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">${v}</urlset>`)
	url_source := []byte(`<url><loc>${v}${v}</loc></url>`)

	var url_block strings.Builder

	for _, page := range PageList {
		url_block.WriteString(string(sub_sprint(url_source, config.Domain, page.URLPath)))
	}

	final := string(sub_sprint(sitemap_source, url_block.String()))

	write_file(filepath.Join(config.Output, "sitemap.xml"), final)
}