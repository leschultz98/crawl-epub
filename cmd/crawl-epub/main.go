package main

import (
	"flag"
	"log"
	"os"
)

const (
	truyenyySource   = "truyenyy"
	metruyencvSource = "metruyencv"
	truyenchuSource  = "truyenchu"
	ttvSource        = "ttv"
)

type config struct {
	source   string
	title    string
	suffix   string
	start    int
	end      int
	startURL string
	endURL   string
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

	flag.StringVar(&cfg.source, "source", "", "ebook sources")
	flag.StringVar(&cfg.title, "title", "", "ebook title")
	flag.StringVar(&cfg.suffix, "suffix", "", "ebook title suffix")
	flag.IntVar(&cfg.start, "start", 1, "start chapter")
	flag.IntVar(&cfg.end, "end", 1, "end chapter")
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
	case truyenchuSource:
		c = &truyenchu{}
	case ttvSource:
		c = &ttv{errorLog: errorLog}
	default:
		log.Fatal("inappropriate source")
	}

	chapters, err := c.getChapters(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = writeEpubs(cfg.title+cfg.suffix, chapters)
	if err != nil {
		log.Fatal(err)
	}
}
