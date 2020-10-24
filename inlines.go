package main

import (
	"regexp"
	"strings"
)

// this file is isolated as a constant
// reminder to replace it with something
// faster

var italics = regexp.MustCompile(`_([^><]+)_`)
var strike  = regexp.MustCompile(`~([^><]+)~`)
var bolds   = regexp.MustCompile(`\*([^><]+)\*`)
var links   = regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)
var code    = regexp.MustCompile("`(.+?)`")

var code_links = regexp.MustCompile(`!\[(.+?)\]\((.+?)\)`)

var inline = regexp.MustCompile(`c\.(.+?){(.+?)}`)

func inlines(v string) string {
	input := []byte(v)

	input = code.ReplaceAll(input,    []byte(`<code>$1</code>`))
	input = links.ReplaceAll(input,   []byte(`<a href='$2'>$1</a>`))
	input = bolds.ReplaceAll(input,   []byte(`<b>$1</b>`))
	input = italics.ReplaceAll(input, []byte(`<i>$1</i>`))
	input = strike.ReplaceAll(input,  []byte(`<s>$1</s>`))

	return string(input)
}

func strip_inlines(v string) string {
	input := []byte(v)

	input = code.ReplaceAll(input,    []byte(`$1`))
	input = links.ReplaceAll(input,   []byte(`$1`))
	input = bolds.ReplaceAll(input,   []byte(`$1`))
	input = italics.ReplaceAll(input, []byte(`$1`))

	return string(input)
}

func inline_code_sub(v string) string {
	v = strings.ReplaceAll(v, `&`, `&amp;`)
	v = strings.ReplaceAll(v, `<`, `&lt;`)
	v = strings.ReplaceAll(v, `>`, `&gt;`)

	input := []byte(v)

	input = code_links.ReplaceAll(input, []byte(`<a href='$2'>$1</a>`))
	input = inline.ReplaceAll(input, []byte(`<span class='token $1'>$2</span>`))

	return string(input)
}