package main

import (
	"fmt"
	"path/filepath"
	"github.com/robertkrimen/otto"
)

var secondary_renders = make(map[string]*Page)

func do_functions(page *Page) {
	for _, f := range page.List.Tokens {
		if f.Type == FUNCTION {
			f.Text = do_single_function(page, f.Text)
		}
	}
}

func do_single_function(page *Page, name string) string {
	path := filepath.Join("_data/functions", name + ".js")

	if !file_exists(path) {
		warning(page.ID + `: external function "` + name + `" does not exist`)
		return ""
	}

	file := string(load_file_bytes(path))

	// new js instance
	vm := otto.New()

	// register current page data into instance
	page_data, _ := vm.Object(`page = {}`)
	page_data.Set("Vars",   page.Vars)
	page_data.Set("Tokens", page.List.Tokens)

	// register project data into instance
	vm.Set("project", config)

	// register page_list
	vm.Set("get_page", func(call otto.FunctionCall) otto.Value {
		id := call.Argument(0).String()

		if p, ok := PageList[id]; ok {
			js_p, err := vm.ToValue(p)

			if err != nil {
				panic(err)
			}

			DepTree[page.ID] = append(DepTree[page.ID], id)

			return js_p
		}

		return otto.Value{}
	})

	// inject Token_Type enums
	token_data, _ := vm.Object(`token_type = {}`)

	for n, str := range token_names {
		token_data.Set(str, n)
	}

	// execute instance
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