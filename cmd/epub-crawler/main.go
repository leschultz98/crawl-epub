package main

import (
	"flag"
	"log"
	"strings"

	"crawl-epub/cmd/epub-crawler/metruyencv"
	"crawl-epub/cmd/epub-crawler/tangthuvien"
	"crawl-epub/cmd/epub-crawler/truyenchu"
	"crawl-epub/internal/epub"
)

const (
	metruyencvHost  = "metruyencv.com"
	tangthuvienHost = "tangthuvien.vn"
	truyenchuHost   = "truyenchu.vn"
)

type crawler interface {
	GetEbook() (string, []*epub.Chapter, error)
}

func main() {
	var url string
	flag.StringVar(&url, "url", "", "url of the start chapter")
	flag.Parse()

	url = strings.TrimPrefix(url, "https://")

	urlParts := strings.Split(url, "/")
	host := urlParts[0]
	paths := urlParts[1:]

	var c crawler
	switch {
	case strings.Contains(host, metruyencvHost):
		c = metruyencv.New(paths)
	case strings.Contains(host, tangthuvienHost):
		c = tangthuvien.New(paths)
	case strings.Contains(host, truyenchuHost):
		c = truyenchu.New(paths)
	}

	title, chapters, err := c.GetEbook()
	if err != nil {
		log.Fatal(err)
	}

	err = epub.WriteEpub(title, chapters)
	if err != nil {
		log.Fatal(err)
	}
}
