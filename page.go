package main

import "strings"

type Page struct {
	ID      string
	URLPath string

	Style  []string
	Script []string

	CurrentParent *Page // @hack

	IsDraft     bool
	HasFunction bool

	Plate *Plate
	List  *Token_List

	Deps map[string]bool
	Vars map[string]string
	Meta map[string]string
}

func create_page(info *File) {
	new_page := &Page {
		ID: info.ID,
	}

	new_page.Vars = make(map[string]string, 8)
	new_page.Meta = make(map[string]string, 8)

	new_page.Vars["page_path"] = new_page.URLPath

	if info.ID == "index" {
		new_page.URLPath = ""
	} else {
		new_page.URLPath = "/" + strings.Replace(info.ID, "/index", "", 1)
	}

	// parse
	bytes := load_file_bytes(info.SourcePath)
	new_page.List = parser(new_page, bytes)

	info.IsDraft = new_page.IsDraft

	info.Page = new_page
}

func get_page(id string) (*Page, bool) {
	if p, ok := AllFiles[id]; ok && p.Type == MARKUP {
		return p.Page, true
	}
	return nil, false
}
