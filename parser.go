package main

import (
	"fmt"
	"bytes"
	"strings"
	"unicode"
)

var DepTree = make(map[string][]string)

type Token_Type int

const (
	tok_begin Token_Type = iota

	H1
	H2
	H3
	H4
	H5
	H6

	tok_headings

	QUOTE
	PARAGRAPH
	LIST_ENTRY

	tok_inline_format

	ERROR
	IMAGE
	TOKEN
	MEDIA
	IMPORT
	SNIPPET
	DIVIDER
	FUNCTION
	BLOCK_CODE
	BLOCK_START
	BLOCK_CLOSE
	CODE_GUTS
	HTML_SNIPPET

	tok_if_statements

	IF_SCOPE_PROJECT
	IF_SCOPE_PROJECT_NOT
	IF_SCOPE_PARENT
	IF_SCOPE_PARENT_NOT
	IF_SCOPE_PAGE
	IF_SCOPE_PAGE_NOT
)

type Token struct {
	Type Token_Type
	Text string
	Line int
	Vars map[string]string
}

var token_names = [...]string {
	"tok_begin",

	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",

	"tok_headings",

	"quote",
	"paragraph",
	"list_entry",

	"tok_inline_format",

	"error",
	"image",
	"token",
	"video",
	"import",
	"snippet",
	"divider",
	"function",
	"block_code",
	"block_start",
	"block_close",
	"code_guts",
	"html_snippet",

	"tok_if_statements",

	"if_scope_project",
	"if_scope_project_not",
	"if_scope_parent",
	"if_scope_parent_not",
	"if_scope_page",
	"if_scope_page_not",
}

func (t Token_Type) String() string {
	return token_names[t]
}



type Token_List struct {
	Tokens []*Token
	Pos    int
	IsCommittable bool
}

func (tree *Token_List) Peek() *Token {
	return tree.Tokens[tree.Pos-1]
}

func (tree *Token_List) Lookahead() *Token {
	if tree.Pos == len(tree.Tokens) {
		return nil
	}
	return tree.Tokens[tree.Pos]
}

func (tree *Token_List) Next() *Token {
	if tree.Pos == len(tree.Tokens) {
		return nil
	}
	tree.Pos++
	return tree.Peek()
}

func (tree *Token_List) Reset() {
	tree.Pos = 0
}



func simple_oko_token(input []rune, r rune) ([]rune, []rune, bool) {
	if input[0] == r {
		input = input[1:]
		input = consume_whitespace(input)

		text := extract_to_newline(input)
		input = input[len(text):]

		return text, input, true
	}
	return nil, nil, false
}

func compare_arbitrary_runes(input []rune, compare string) (int, bool) {
	test := []rune(compare)
	for i, r := range test {
		if r != input[i] {
			return 0, false
		}
	}
	return len(test), true
}

func extract_identifier(input []rune) []rune {
	var extract []rune
	for _, r := range input {
		if !(unicode.IsLetter(r) || r == '_' || r == '.') { // @todo variable semantics
			return extract
		}
		extract = append(extract, r)
	}
	return nil
}

func extract_to_newline(input []rune) []rune {
	return input[0:jump_to_next_newline(input)]
}

func jump_to_next_newline(input []rune) int {
		c := 0
		for _, r := range input {
			if r == '\n' || r == '\r' {
				return c
			}
			c++
		}
		return c
}

func jump_to_next_char(input []rune, test rune) int {
	c := 0
	for _, r := range input {
		if r == test {
			return c
		}
		c++
	}
	return c
}

func consume_whitespace(input []rune) []rune {
	for i, r := range input {
		if !unicode.IsSpace(r) {
			return input[i:]
		}
	}
	return input[len(input):]
}

func count_newlines(input []rune) int {
	c := 0
	for _, r := range input {
		if r == '\n' || r == '\r' {
			c++
		}
	}
	return c + 1
}

func count_sequential_runes(input []rune, check rune) int {
	c := 0
	for _, r := range input {
		if r != check {
			return c
		}
		c++
	}
	return c
}


// block helpers
func pop(a []*Token) []*Token {
	if len(a) <= 0 {
		return a
	}
	return a[0:len(a)-1]
}

func get(a []*Token) (*Token, bool) {
	if len(a) > 0 {
		return a[len(a)-1], true
	}
	return nil, false
}



// @note block open/close balance checking pass?
func parser(page *Page, source []byte) *Token_List {
	input := bytes.Runes(source)

	total_lines := count_newlines(input)

	line_no := func(input []rune) int {
		return total_lines - count_newlines(input) + 1
	}

	// must be re-rendered every call due to relations
	committable := true

	var list []*Token
	var active_block []*Token

	for len(input) > 0 {
		input = consume_whitespace(input)

		if len(input) == 0 {
			break
		}

		if input[0] == '}' {
			input = input[1:]

			list = append(list, &Token{BLOCK_CLOSE, "", line_no(input), nil})

			active_block = pop(active_block)

			continue
		}

		// single-line comments
		if c := count_sequential_runes(input, '/'); c >= 2 {
			input = input[jump_to_next_newline(input):]
			continue
		}

		if input[0] == '#' {
			c     := count_sequential_runes(input, '#')
			input  = consume_whitespace(input[c:])
			text  := extract_to_newline(input)
			input  = input[len(text):]

			var heading Token_Type

			switch c {
				case 1: heading = H1
				case 2: heading = H2
				case 3: heading = H3
				case 4: heading = H4
				case 5: heading = H5
				case 6: heading = H6
			}

			list = append(list, &Token{heading, string(text), line_no(input), nil})
			continue
		}

		if text, update_input, ok := simple_oko_token(input, '%'); ok {
			input = update_input
			list = append(list, &Token{IMAGE, string(text), line_no(input), nil})
			continue
		}
		if text, update_input, ok := simple_oko_token(input, '@'); ok {
			input = update_input
			list = append(list, &Token{MEDIA, string(text), line_no(input), nil})
			continue
		}
		if text, update_input, ok := simple_oko_token(input, '+'); ok {
			input  = update_input
			t     := string(text)
			name  := strings.SplitN(t, " ", 2)[0]

			DepTree[name] = append(DepTree[name], page.ID)

			list = append(list, &Token{IMPORT, t, line_no(input), nil})
			continue
		}
		if text, update_input, ok := simple_oko_token(input, '>'); ok {
			input  = update_input
			t     := string(text)
			name  := "snip_" + t

			DepTree[name] = append(DepTree[name], page.ID)

			list = append(list, &Token{SNIPPET, t, line_no(input), nil})
			continue
		}
		if text, update_input, ok := simple_oko_token(input, '&'); ok {
			input = update_input
			list = append(list, &Token{TOKEN, string(text), line_no(input), nil})
			continue
		}
		if text, update_input, ok := simple_oko_token(input, 'ø'); ok {
			input = update_input
			list = append(list, &Token{FUNCTION, string(text), line_no(input), nil})
			continue
		}
		if text, update_input, ok := simple_oko_token(input, '$'); ok {
			input = update_input
			list = append(list, &Token{QUOTE, string(text), line_no(input), nil})
			continue
		}
		if text, update_input, ok := simple_oko_token(input, '.'); ok {
			input = update_input
			list = append(list, &Token{PARAGRAPH, string(text), line_no(input), nil})
			continue
		}

		// html snippets (preserve leading whitespace - 1)
		if input[0] == '*' {
			input  = input[2:]
			text  := extract_to_newline(input)
			input  = input[len(text):]
			list   = append(list, &Token{HTML_SNIPPET, string(text), line_no(input), nil})
			continue
		}

		if input[0] == '-' {
			if count_sequential_runes(input, '-') == 3 {
				input = input[3:]
				list = append(list, &Token{DIVIDER, "", line_no(input), nil})
				continue
			}

			if text, update_input, ok := simple_oko_token(input, '-'); ok {
				input = update_input
				list = append(list, &Token{LIST_ENTRY, string(text), line_no(input), nil})
				continue
			}
		}

		// variables AND blocks
		ident := extract_identifier(input)

		if len(ident) > 0 {
			test_input := consume_whitespace(input[len(ident):])

			// we are a variable
			if test_input[0] == ':' {
				test_input = consume_whitespace(test_input[1:])

				value := extract_to_newline(test_input)
				k, v  := string(ident), string(value)

				input = test_input[len(value):] // overwrite input w changes

				// check if submap variable
				if strings.Contains(k, ".") {
					n := strings.Split(k, ".")

					switch n[0] {
						case "meta":
							page.Meta[n[1]] = v
							page.Vars[k] = v
					}

					continue
				}

				switch k {
					case "script":
						page.Script = append(page.Script, v)

					/*case "tags":
						for _, t := range strings.Fields(v) {
							page.Tags[t] = true
						}*/

					case "draft":
						if v == "true" {
							page.IsDraft = true
							page.Vars["draft"] = "true"
						}

					default:
						if a, ok := get(active_block); ok {
							a.Vars[k] = v
						} else {
							page.Vars[k] = v
						}
				}

				continue
			}

			c := jump_to_next_newline(test_input)

			// we are a block
			if test_input[c-1] == '{' {
				str_ident := string(ident)

				if str_ident == "code" {
					lang := string(extract_identifier(test_input))

					if len(lang) == 0 {
						lang = "code"
					}

					test_input = test_input[c+1:]
					// subtract from line_no  ^ because we sliced it off just above
					n := line_no(test_input) - 1

					list = append(list, &Token{BLOCK_CODE, lang, n, nil})

					var indent   int
					var count    int
					var last     rune
					var ws_count int

					for _, r := range test_input {
						if r == '\t' || r == ' ' {
							indent++
						} else {
							break
						}
					}

					for _, r := range test_input {
						if r == '}' && unicode.IsSpace(last) {
							break
						}
						last = r
						count++
					}

					content := test_input[0:count]

					for i := len(content); i > 0; i-- {
						if !unicode.IsSpace(content[i-1]) {
							break
						}
						ws_count++
					}

					content = content[0:count-ws_count]

					code := string(content)

					// @hack replace me
					code  = strings.ReplaceAll(code, "\n" + strings.Repeat("\t", indent), "\n")[indent:]

					code  = strings.ReplaceAll(code, "\t", "    ")
					code  = strings.ReplaceAll(code, "\\}", "}")

					list = append(list, &Token{CODE_GUTS, code, n+1, nil})

					input = test_input[count+1:]

					continue
				}

				// if statement
				if str_ident == "if" {
					if_token := Token{}
					if_input := test_input

					found_valid_scope := false
					is_not := false

					if if_input[0] == '!' {
						if_input = if_input[1:]
						is_not   = true
					}

					if count, ok := compare_arbitrary_runes(if_input, "project"); ok {
						if_input = consume_whitespace(if_input[count:])
						found_valid_scope = true

						if is_not {
							if_token.Type = IF_SCOPE_PROJECT_NOT
						} else {
							if_token.Type = IF_SCOPE_PROJECT
						}

					} else if count, ok := compare_arbitrary_runes(if_input, "parent"); ok {
						if_input = consume_whitespace(if_input[count:])
						found_valid_scope = true

						committable = false

						if is_not {
							if_token.Type = IF_SCOPE_PARENT_NOT
						} else {
							if_token.Type = IF_SCOPE_PARENT
						}

					} else if count, ok := compare_arbitrary_runes(if_input, "page"); ok {
						if_input = consume_whitespace(if_input[count:])
						found_valid_scope = true

						if is_not {
							if_token.Type = IF_SCOPE_PAGE_NOT
						} else {
							if_token.Type = IF_SCOPE_PAGE
						}
					}

					if !found_valid_scope {
						ident := extract_identifier(if_input)
						list = append(list, &Token{ERROR, "no such scope " + string(ident), line_no(if_input), nil})
						continue
					}

					if if_input[0] == '.' {
						if_input = if_input[1:]
					} else {
						list = append(list, &Token{ERROR, "missing '.' separator in if-statement", line_no(if_input), nil})
						continue
					}

					ident := extract_identifier(if_input)

					if len(ident) > 0 {
						if_input = if_input[len(ident):]
					} else {
						list = append(list, &Token{ERROR, "no variable in if-statement", line_no(if_input), nil})
						continue
					}

					if_token.Text = string(ident)
					if_token.Line = line_no(if_input)
					if_token.Vars = make(map[string]string)

					list = append(list, &if_token)
					active_block = append(active_block, &if_token)

				} else {
					b := &Token{BLOCK_START, str_ident, line_no(test_input), nil}
					b.Vars = make(map[string]string)
					list = append(list, b)
					active_block = append(active_block, b)
				}

				input = test_input[c:] // replace changes

				continue
			}
		}

		// PARAGRAPH
		text := extract_to_newline(input)
		list  = append(list, &Token{PARAGRAPH, string(text), line_no(input), nil})
		input = input[len(text):]
	}

	if name, ok := page.Vars["plate"]; ok {
		n := "plate_" + name
		DepTree[n] = append(DepTree[n], page.ID)

		plate := load_plate(name)

		if len(plate.SnippetBefore) > 0 {
			for _, s := range plate.SnippetBefore {
				name := "snip_" + s
				DepTree[name] = append(DepTree[name], page.ID)
			}
		}

		if len(plate.SnippetAfter) > 0 {
			for _, s := range plate.SnippetAfter {
				name := "snip_" + s
				DepTree[name] = append(DepTree[name], page.ID)
			}
		}
	}

	return &Token_List{Tokens: list, IsCommittable:committable}
}

// dev
func print_syntax_tree(list *Token_List) {
	for _, entry := range list.Tokens {
		fmt.Println(entry)
	}
}