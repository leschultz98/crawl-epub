package main

import (
	"flag"
	"log"
	"strings"

	"crawl-epub/internal/crawlers"
	"crawl-epub/internal/epub"
)

func main() {
	var url string
	flag.StringVar(&url, "url", "", "url of the start chapter")
	flag.Parse()

	url = strings.TrimPrefix(url, "https://")

	urlParts := strings.Split(url, "/")
	host := urlParts[0]
	paths := urlParts[1:]

	c := crawlers.GetCrawler(host, paths)

	title, chapters, err := c.GetEbook()
	if err != nil {
		log.Fatal(err)
	}

	err = epub.WriteEpub(title, chapters)
	if err != nil {
		log.Fatal(err)
	}
}
