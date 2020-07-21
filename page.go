package main

import (
	"strings"
	"path/filepath"
)

var PageList = make(map[string]*Page)

type Page struct {
	ID         string
	SourcePath string
	OutputPath string
	URLPath    string

	Style      []string
	Script     []string

	CurrentParent *Page // @hack

	IsDraft    bool
	IsIndex    bool
	Format     File_Format

	Plate      *Plate
	List       *Token_List

	Vars       map[string]string
	Meta       map[string]string
	Tags       map[string]bool
}

func make_page(info *File_Info) *Page {
	if p, exists := PageList[info.ID]; exists {
		return p
	}

	output := filepath.Join(config.Output, info.ID + ".html")

	new_page := &Page{
		ID: info.ID,
		SourcePath: info.Path,
		OutputPath: output,
		Format: info.Format,
	}

	new_page.Vars = make(map[string]string, 8)
	new_page.Meta = make(map[string]string, 8)
	new_page.Tags = make(map[string]bool,   8)

	if info.ID == "index" {
		new_page.URLPath = ""
		new_page.IsIndex = true
	} else {
		new_page.URLPath = "/" + strings.Replace(info.ID, "/index", "", 1)
		new_page.IsIndex = false
	}

	new_page.Vars["page_path"] = new_page.URLPath

	PageList[info.ID] = new_page

	return new_page
}