package main

import "regexp"

// this file is isolated as a constant
// reminder to replace it with something
// faster

var italics = regexp.MustCompile(`_(.+?)_`)
var bolds   = regexp.MustCompile(`\*(.+?)\*`)
var links   = regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)
var code    = regexp.MustCompile("`(.+?)`")

func inlines(v string) string {
	input := []byte(v)

	input = links.ReplaceAll(input,   []byte(`<a href='$2'>$1</a>`))
	input = bolds.ReplaceAll(input,   []byte(`<b>$1</b>`))
	input = italics.ReplaceAll(input, []byte(`<i>$1</i>`))
	input = code.ReplaceAll(input,    []byte(`<code>$1</code>`))

	return string(input)
}

func strip_inlines(v string) string {
	input := []byte(v)

	input = links.ReplaceAll(input,   []byte(`$1`))
	input = bolds.ReplaceAll(input,   []byte(`$1`))
	input = italics.ReplaceAll(input, []byte(`$1`))
	input = code.ReplaceAll(input,    []byte(`$1`))

	return string(input)
}