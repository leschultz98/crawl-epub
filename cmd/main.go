package main

import (
	"fmt"
	"log"

	"github.com/bmaupin/go-epub"
)

const (
	urlF        = "http://localhost:5001/truyen/ai-con-khong-la-cai-nguoi-tu-hanh-roi/chuong-%d.html"
	title       = "book"
	fontPath    = "../assets/fonts/OpenSans-Regular.ttf"
	cssPath     = "../assets/styles/styles.css"
	outputPathF = "../ebooks/%s.epub"
)

func main() {
	e := epub.NewEpub(title)

	_, err := e.AddFont(fontPath, "")
	if err != nil {
		log.Fatal(err)
	}

	css, err := e.AddCSS(cssPath, "")
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for {
		count++
		chapter, err := getChapter(fmt.Sprintf(urlF, count))
		if err != nil {
			log.Fatal(err)
		}

		if chapter.content == "" {
			break
		}

		e.AddSection(chapter.content, chapter.title, "", css)
	}

	err = e.Write(fmt.Sprintf(outputPathF, title))
	if err != nil {
		log.Fatal(err)
	}
}
