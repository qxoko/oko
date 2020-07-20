// +build scripting

package main

import (
	"fmt"
	"path/filepath"
	"github.com/robertkrimen/otto"
)

func do_script(page *Page, name string) string {
	path := filepath.Join("_data/functions", name + ".js")

	if !file_exists(path) {
		warning(page.ID + `: external function "` + name + `" does not exist`)
		return ""
	}

	file := string(load_file_bytes(path))

	vm := otto.New()

	page_data, _ := vm.Object(`page = {}`)

	page_data.Set("Vars",   page.Vars)
	page_data.Set("Tokens", page.List.Tokens)

	vm.Set("project", config)

	_, err := vm.Run(file)

	if err != nil {
		fmt.Println(err)
	}

	value, err := vm.Get("result")

	if err != nil {
		panic(err)
	}

	str, _ := value.Export() // this err is always nil in otto

	if str == nil {
		return ""
	}

	return str.(string)
}