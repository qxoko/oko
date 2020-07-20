package main

import "os"

func make_project() {
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

	os.MkdirAll(`_data/plates`,    os.ModePerm)
	os.MkdirAll(`_data/snippets`,  os.ModePerm)
	os.MkdirAll(`_data/functions`, os.ModePerm)

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