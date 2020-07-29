package main

import (
	"strings"
	"unicode"
)

// turns any string into a clean HTML ID attribute
func make_element_id(source string) string {
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

// takes a text block containing ${variables}
// and remaps them against a go map
// hard argument determines whether unmatched
// variables are left in the text
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

// substitutes all matches of a specific
// variable in text with new string
func sub(source, r, v string) string {
	return strings.ReplaceAll(source, `${` + r + `}`, v)
}

// slightly faster sub that only matches the
// %s convention
func sub_content(source, v string) string {
	return strings.ReplaceAll(source, `%s`, v)
}

// extremely fast sprintf for %s convention
func sub_sprint(source string, v ...string) string {
	for _, x := range v {
		source = strings.Replace(source, `%s`, x, 1)
	}
	return source
}