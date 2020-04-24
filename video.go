package main

import (
	"fmt"
	"strings"
	"unicode"
	"strconv"
)

type Video_Service int

const (
	YOUTUBE Video_Service = iota
	VIMEO
)

func vimeo(viewcode, ratio string, args []string) string {
	iframe := sub_sprint(`<div class="video"><div class="video-container"${v}><iframe src="https://player.vimeo.com/video/${v}?color=0&title=0&byline=0&portrait=0" frameborder="0" allow="fullscreen" allowfullscreen></iframe></div></div>`, ratio, viewcode)

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

func youtube(viewcode, ratio string, args []string) string {
	iframe := sub_sprint(`<div class="video"><div class="video-container"${v}><iframe src="https://www.youtube-nocookie.com/embed/${v}?rel=0&controls=1" frameborder="0" allow="accelerometer; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe></div></div>`, ratio, viewcode)

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

func video(s string) string {
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

	if strings.Contains(args[1], ":") {
		v := strings.SplitN(args[1], ":", 2)

		x, err := strconv.ParseFloat(v[0], 32); if err != nil { panic(err) }
		y, err := strconv.ParseFloat(v[1], 32); if err != nil { panic(err) }

		ratio = fmt.Sprintf(` style="padding-top: %.2f%%"`, y / x * 100.0)
	}

	switch service {
		case YOUTUBE: return youtube(viewcode, ratio, args[1:])
		case VIMEO:   return vimeo(viewcode,   ratio, args[1:])
	}

	return ""
}