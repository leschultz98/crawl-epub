package main

import (
	"flag"
	"log"
)

const truyenyySource = "truyenyy"

type config struct {
	source         string
	from           int
	end            int
	title          string
	chapterListUrl string
}

type chapter struct {
	title   string
	content string
}

type crawler interface {
	getChapters(config) ([]*chapter, error)
}

func main() {
	var cfg config

	flag.StringVar(&cfg.source, "source", truyenyySource, "ebook sources: truyenyy, ttv")
	flag.StringVar(&cfg.title, "title", "ai-con-khong-la-cai-nguoi-tu-hanh-roi", "ebook title")
	flag.IntVar(&cfg.from, "from", 1, "chapter start")
	flag.IntVar(&cfg.end, "end", 1, "chapter end")
	flag.Parse()

	var c crawler

	switch cfg.source {
	case truyenyySource:
		c = truyenyy{}

	default:
		log.Fatal("inappropriate source")
	}

	chapters, err := c.getChapters(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = writeEpub(cfg.title, chapters)
	if err != nil {
		log.Fatal(err)
	}
}
