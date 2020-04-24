package main

import (
	"os"
	"fmt"
)

var Warnings []string

func warning(msg string) {
	Warnings = append(Warnings, msg)
}

func print_warnings() {
	if len(Warnings) == 0 {
		return
	}

	fmt.Println("[ø] warnings\n")

	for _, msg := range Warnings {
		fmt.Println("   ", msg)
	}

	fmt.Println()
}

func render_error(the_page *Page, tok *Token, msg string) {
	fmt.Printf("%s L%d: %s %q\n", the_page.ID, tok.Line, msg, tok.Text)
	os.Exit(1)
}