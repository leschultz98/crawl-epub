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
	source         string
	from           int
	end            int
	length         int
	title          string
	chapterListUrl string
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
	flag.IntVar(&cfg.from, "from", 1, "chapter start")
	flag.IntVar(&cfg.end, "end", 0, "chapter end (require for truyenyy")
	flag.StringVar(&cfg.chapterListUrl, "list", "https://m.truyen.tangthuvien.vn/danh-sach-chuong/13450", "chapter list url")
	flag.Parse()

	cfg.length = cfg.end - cfg.from + 1

	var c crawler

	switch cfg.source {
	case truyenyySource:
		c = truyenyy{}
	case ttvSource:
		c = ttv{}
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
