package main

import (
	"fmt"
	"strings"
	"unicode"
	"strconv"
	"path/filepath"
)

// adds media_prefix_path (oko.json) to
// images unless image is external or already
// has prefix
func image_checker(v string) string {
	if !strings.HasPrefix(v, `http`) && !strings.HasPrefix(v, config.ImagePrefix) {
		v = config.ImagePrefix + v
	}
	return v
}

func make_favicon(f string) string {
	var tag string

	switch filepath.Ext(f) {
		case ".ico": tag = `<link rel='icon' type='image/x-icon' href='%s'>`
		case ".png": tag = `<link rel='icon' type='image/png' href='%s'>`
		case ".gif": tag = `<link rel='icon' type='image/gif' href='%s'>`
		default: panic("bad favicon format")
	}

	return sub_content(tag, f)
}

//
// External Services
//
type Media_Service int

const (
	YOUTUBE Media_Service = iota
	VIMEO
)

func media_vimeo(viewcode, ratio string, args []string) string {
	iframe := sub_sprint(`<div class='video'><div class='video-container'%s><iframe src='https://player.vimeo.com/video/%s?color=0&title=0&byline=0&portrait=0' frameborder='0' allow='fullscreen' allowfullscreen></iframe></div></div>`, ratio, viewcode)

	if len(args) == 0 {
		return iframe
	}

	for _, a := range args {
		if a[0] == '#' {
			iframe = strings.Replace(iframe, `color=0`, `color=` + a[1:len(a)], 1)
			continue
		}

		switch a {
			case "hide_all":
				iframe = strings.Replace(iframe, `&title=0&byline=0&portrait=0`, ``, 1)

			case "hide_title":
				iframe = strings.Replace(iframe, `&title=0`, ``, 1)

			case "hide_portrait":
				iframe = strings.Replace(iframe, `&portrait=0`, ``, 1)

			case "hide_byline":
				iframe = strings.Replace(iframe, `&byline=0`, ``, 1)
		}
	}

	return iframe
}

func media_youtube(viewcode, ratio string, args []string) string {
	iframe := sub_sprint(`<div class='video'><div class='video-container'%s><iframe src='https://www.youtube-nocookie.com/embed/%s?rel=0&controls=1' frameborder='0' allow='accelerometer; encrypted-media; gyroscope; picture-in-picture' allowfullscreen></iframe></div></div>`, ratio, viewcode)

	if len(args) == 0 {
		return iframe
	}

	for _, a := range args {
		switch a {
			case "hide_controls":
				iframe = strings.Replace(iframe, `&controls=1`, `&controls=0`, 1)
		}
	}

	return iframe
}

func media(s string) string {
	if s == "" {
		return ""
	}

	args     := strings.Split(s, " ")
	viewcode := args[0]

	service  := VIMEO
	ratio    := ""

	for _, r := range viewcode {
		if unicode.IsLetter(r) {
			service = YOUTUBE
			break
		}
	}

	if len(args) > 1 {
		if strings.Contains(args[1], ":") {
			v := strings.SplitN(args[1], ":", 2)

			x, err := strconv.ParseFloat(v[0], 32); if err != nil { panic(err) }
			y, err := strconv.ParseFloat(v[1], 32); if err != nil { panic(err) }

			ratio = fmt.Sprintf(` style="padding-top: %.2f%%"`, y / x * 100.0)
		}

		args = args[1:]
	} else {
		args = nil
	}

	switch service {
		case YOUTUBE: return media_youtube(viewcode, ratio, args)
		case VIMEO:   return media_vimeo(viewcode,   ratio, args)
	}

	return ""
}