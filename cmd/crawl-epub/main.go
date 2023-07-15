package main

import (
	"flag"
	"log"
)

const (
	truyenyySource = "truyenyy"
	ttvSource      = "ttv"
)

type config struct {
	source string
	length int
	title  string
	bookID string
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

	flag.StringVar(&cfg.source, "source", ttvSource, "ebook sources: truyenyy, ttv")
	flag.StringVar(&cfg.title, "title", "trafford-nguoi-mua-cau-lac-bo", "ebook title")
	flag.IntVar(&cfg.length, "length", 0, "number of chapter (truyenyy: from start, ttv: to end)")
	flag.StringVar(&cfg.bookID, "bookID", "13450", "ttv book id")
	flag.Parse()

	var c crawler

	switch cfg.source {
	case truyenyySource:
		c = &truyenyy{}
	case ttvSource:
		c = &ttv{}
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
