package main

import (
	"os"
	"strings"
	"encoding/json"
)

type Config struct {
	Domain  string
	Output  string
	Favicon string
	Title   string

	StyleRender string

	DoAllPages      bool `json:"do_all_pages"`
	ShowDrafts      bool `json:"show_drafts"`

	DoCodeHighlight bool `json:"code_highlight"`
	Sitemap bool

	Style      []string
	Include    []string
	Extensions []string

	ImagePrefix string `json:"image_path_prefix"`

	Meta map[string]string
	Vars map[string]string
}

func load_config() *Config {
	var config Config

	p := "_data/oko.json"

	if !file_exists(p) {
		return nil
	}

	err := json.Unmarshal(load_file_bytes(p), &config)

	if err != nil {
		panic(err)
	}

	if config.Domain == "" {
		panic("no domain name in _data/oko.json!")
	}

	if !strings.HasPrefix(config.Domain, "https://") {
		config.Domain = "https://" + config.Domain
	}

	if config.Output == "" {
		config.Output = "public" // set a reasonable default
	}

	if config.Favicon != "" {
		config.Favicon = make_favicon(config.Favicon)
	}

	if config.Vars == nil {
		config.Vars = make(map[string]string, 8)
	}

	if len(config.Extensions) == 0 {
		config.Extensions = []string{`.ø`, `.html`}
	} else {
		has_html := false

		for _, e := range config.Extensions {
			if e == `.html` {
				has_html = true
				break
			}
		}

		if !has_html {
			config.Extensions = append(config.Extensions, `.html`)
		}
	}

	config.StyleRender = render_style(config.Style, ``)

	return &config
}

func make_new_config_file() {
	template := []byte(`{
	"domain": "https://website.com",
	"favicon": "/favicon.png",
	"sitemap": true,
	"image_path_prefix": "",
	"style": [],
	"include": [],
	"meta": {
		"twitter_creator": "@jack",
		"image": "default_card.png",
		"description": "Default description for search engines and embeds"
	}
}`)

	f, err := os.Create(`_data/oko.json`)
	defer f.Close()

	if err != nil {
		panic(err)
	}

	_, err = f.Write(template)

	if err != nil {
		panic(err)
	}
}