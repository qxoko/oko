package main

func project(tok *Token) bool {
	switch tok.Text {
		case "domain":
			if config.Domain != "" {
				return true
			}
		case "output":
			if config.Output != "" {
				return true
			}
		case "favicon":
			if config.Favicon != "" {
				return true
			}
		case "style":
			if len(config.Style) > 0 {
				return true
			}
		case "include":
			if len(config.Include) > 0 {
				return true
			}
		case "meta":
			if len(config.Meta) > 0 {
				return true
			}
	}
	return false
}

func page(the_page *Page, tok *Token) bool {
	switch tok.Text {
		case "style":
			if len(the_page.Style) > 0 {
				return true
			}
		case "script":
			if len(the_page.Script) > 0 {
				return true
			}
		/*case "tags":
			if len(the_page.Tags) > 0 {
				return true
			}*/
		default:
			if v, ok := the_page.Vars[tok.Text]; ok {
				if v != "false" {
					return true
				}
				return false
			}
	}
	return false
}

func check_if_statement(the_page *Page, tok *Token) bool {
	switch tok.Type {
		case IF_SCOPE_PROJECT:     return project(tok)
		case IF_SCOPE_PROJECT_NOT: return !project(tok)
		case IF_SCOPE_PAGE:        return page(the_page, tok)
		case IF_SCOPE_PAGE_NOT:    return !page(the_page, tok)
		case IF_SCOPE_PARENT:      return page(the_page.CurrentParent, tok)
		case IF_SCOPE_PARENT_NOT:  return !page(the_page.CurrentParent, tok)
	}
	return false
}