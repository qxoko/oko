package main

import (
	"bytes"
	"bufio"
	"regexp"
	"strings"
	"encoding/json"
)

var SyntaxList = make(map[string]*Highlighter)

type Highlighter_Data struct {

	// types
	String  []string
	Entity  []string
	Builtin []string
	Keyword []string
	Number  []string
	Boolean []string
	Comment []string
}

type Highlighter struct {

	// types
	String  []*regexp.Regexp
	Entity  []*regexp.Regexp
	Builtin []*regexp.Regexp
	Keyword []*regexp.Regexp
	Number  []*regexp.Regexp
	Boolean []*regexp.Regexp
	Comment []*regexp.Regexp
}

func load_syntax(name string) *Highlighter {
	if h, ok := SyntaxList[name]; ok {
		return h
	}

	var data Highlighter_Data

	path := `_data/syntax/` + name + `.json`

	if file_exists(path) {
		err := json.Unmarshal(load_file_bytes(path), &data)

		if err != nil {
			panic(sub_sprint(`failed to parse JSON in "%s"\nerror: "%s"`, path, err.Error()))
		}
	} else {
		panic(`no such syntax file ` + name)
	}

	var compiled Highlighter

	for _, d := range data.String {
		compiled.String = append(compiled.String, regexp.MustCompile(d))
	}
	for _, d := range data.Entity {
		compiled.Entity = append(compiled.Entity, regexp.MustCompile(d))
	}
	for _, d := range data.Builtin {
		compiled.Builtin = append(compiled.Builtin, regexp.MustCompile(d))
	}
	for _, d := range data.Keyword {
		compiled.Keyword = append(compiled.Keyword, regexp.MustCompile(d))
	}
	for _, d := range data.Number {
		compiled.Number = append(compiled.Number, regexp.MustCompile(d))
	}
	for _, d := range data.Boolean {
		compiled.Boolean = append(compiled.Boolean, regexp.MustCompile(d))
	}
	for _, d := range data.Comment {
		compiled.Comment = append(compiled.Comment, regexp.MustCompile(d))
	}

	SyntaxList[name] = &compiled

	return &compiled
}

func highlight_code(text, hlight string) string {
	h := load_syntax(hlight)

	scanner := bufio.NewScanner(strings.NewReader(text))

	var final bytes.Buffer

	for scanner.Scan() {
	    line := scanner.Bytes()

		for _, r := range h.String {
			line = r.ReplaceAll(line, []byte(`<span class='token string'>$1</span>`))
		}
		for _, r := range h.Entity {
			line = r.ReplaceAll(line, []byte(`<span class='token entity'>$1</span>`))
		}
		for _, r := range h.Builtin {
			line = r.ReplaceAll(line, []byte(`<span class='token builtin'>$1</span>`))
		}
		for _, r := range h.Keyword {
			line = r.ReplaceAll(line, []byte(`<span class='token keyword'>$1</span>`))
		}
		for _, r := range h.Number {
			line = r.ReplaceAll(line, []byte(`<span class='token number'>$1</span>`))
		}
		for _, r := range h.Boolean {
			line = r.ReplaceAll(line, []byte(`<span class='token boolean'>$1</span>`))
		}
		for _, r := range h.Comment {
			line = r.ReplaceAll(line, []byte(`<span class='token comment'>$1</span>`))
		}

		final.Write(line)
		final.Write([]byte(`<br>`))
	}

	return final.String()
}