package crawlers

import (
	"errors"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/crawlers/metruyencv"
	"crawl-epub/internal/crawlers/tangthuvien"
	"crawl-epub/internal/crawlers/truyenchu"
	"crawl-epub/internal/epub"
)

type Crawler interface {
	GetEbook() (string, []*epub.Chapter, error)
}

func GetCrawler(host string, cfg *config.Config) (Crawler, error) {
	var c Crawler

	c, e := truyenchu.New(host, cfg)
	if e == nil {
		return c, nil
	}

	c, e = metruyencv.New(host, cfg)
	if e == nil {
		return c, nil
	}

	c, e = tangthuvien.New(host, cfg)
	if e == nil {
		return c, nil
	}

	return nil, errors.New("")
}
