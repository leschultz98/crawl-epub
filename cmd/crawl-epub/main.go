package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/bmaupin/go-epub"
)

const (
	fontPath  = "assets/fonts/OpenSans-Regular.ttf"
	cssPath   = "assets/styles/styles.css"
	outputDir = "ebooks"
)

type application struct {
	client *http.Client
}

func main() {
	var (
		title string
		from  int
		end   int
	)

	flag.StringVar(&title, "title", "ai-con-khong-la-cai-nguoi-tu-hanh-roi", "ebook title")
	flag.IntVar(&from, "from", 1, "ebook from chapter")
	flag.IntVar(&end, "end", 0, "ebook to chapter (require)")
	flag.Parse()

	if end == 0 {
		log.Fatal("must set flag end greater than 0")
	}

	app := application{
		client: &http.Client{
			// disable HTTP/2
			Transport: &http.Transport{
				TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
			},
		},
	}

	length := end - from + 1
	chapters := make([]*chapter, length)

	var wg sync.WaitGroup
	wg.Add(length)
	bar := newBar(length, "[1/3] Get chapter...")

	for i := 0; i < length; i++ {
		go func(i int) {
			number := i + from
			chapter, err := app.getChapter(fmt.Sprintf("https://truyenyy.vip/truyen/%s/chuong-%d.html", title, number), number)
			if err != nil {
				log.Fatal(err)
			}

			chapters[i] = chapter
			bar.Add(1)
			wg.Done()
		}(i)
	}

	wg.Wait()

	e := epub.NewEpub(title)

	_, err := e.AddFont(fontPath, "")
	if err != nil {
		log.Fatal(err)
	}

	css, err := e.AddCSS(cssPath, "")
	if err != nil {
		log.Fatal(err)
	}

	bar = newBar(length, "[2/3] Write chapter...")
	for i := range chapters {
		e.AddSection(chapters[i].content, chapters[i].title, "", css)
		bar.Add(1)
	}

	bar = newBar(-1, "[3/3] Write epub...")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	err = e.Write(fmt.Sprintf("%s/%s.epub", outputDir, title))
	if err != nil {
		log.Fatal(err)
	}
	bar.Finish()
}
