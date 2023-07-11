package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bmaupin/go-epub"
)

const (
	fontPath  = "assets/fonts/OpenSans-Regular.ttf"
	cssPath   = "assets/styles/styles.css"
	outputDir = "ebooks"
)

func main() {
	var (
		title, urlF string
		from        int
	)

	flag.StringVar(&title, "title", "ai-con-khong-la-cai-nguoi-tu-hanh-roi", "ebook title")
	flag.StringVar(&urlF, "url", "http://localhost:5001/truyen/ai-con-khong-la-cai-nguoi-tu-hanh-roi/chuong-%d.html", "ebook url pattern")
	flag.IntVar(&from, "from", 1, "ebook from chapter")
	flag.Parse()

	e := epub.NewEpub(title)

	_, err := e.AddFont(fontPath, "")
	if err != nil {
		log.Fatal(err)
	}

	css, err := e.AddCSS(cssPath, "")
	if err != nil {
		log.Fatal(err)
	}

	for {
		chapter, err := getChapter(fmt.Sprintf(urlF, from))
		if err != nil {
			log.Fatal(err)
		}

		if chapter.content == "" {
			break
		}

		e.AddSection(chapter.content, chapter.title, "", css)
		from++
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	err = e.Write(fmt.Sprintf("%s/%s.epub", outputDir, title))
	if err != nil {
		log.Fatal(err)
	}
}
