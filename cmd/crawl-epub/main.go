package main

import (
	"flag"
	"log"
	"os"
)

const (
	truyenyySource   = "truyenyy"
	metruyencvSource = "metruyencv"
	ttvSource        = "ttv"
)

type config struct {
	source   string
	length   int
	startURL string
	endURL   string
	title    string
	bookID   string
}

type chapter struct {
	title   string
	content string
	url     string
}

type crawler interface {
	getChapters(*config) ([]*chapter, error)
}

func main() {
	var cfg config

	flag.StringVar(&cfg.source, "source", "", "ebook sources: truyenyy, metruyencv, ttv")
	flag.StringVar(&cfg.title, "title", "", "ebook title")
	flag.IntVar(&cfg.length, "length", 0, "number of chapter")
	flag.StringVar(&cfg.startURL, "startURL", "", "start chapter url")
	flag.StringVar(&cfg.endURL, "endURL", "", "end chapter url")
	flag.StringVar(&cfg.bookID, "bookID", "", "ttv book id")
	flag.Parse()

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var c crawler

	switch cfg.source {
	case truyenyySource:
		c = &truyenyy{}
	case metruyencvSource:
		c = &metruyencv{}
	case ttvSource:
		c = &ttv{errorLog: errorLog}
	default:
		log.Fatal("inappropriate source")
	}

	chapters, err := c.getChapters(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = writeEpub(cfg.title, chapters)
	if err != nil {
		log.Fatal(err)
	}
}
