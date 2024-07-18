package crawlers

import (
	"errors"
	"strings"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/crawlers/metruyencv"
	"crawl-epub/internal/crawlers/tangthuvien"
	"crawl-epub/internal/crawlers/truyenchu"
	"crawl-epub/internal/epub"
)

const (
	truyenchuHost = "truyenchu.vn"
)

type Crawler interface {
	GetEbook() (string, []*epub.Chapter, error)
}

func GetCrawler(host string, cfg *config.Config) (Crawler, error) {
	var c Crawler

	switch {
	case strings.Contains(host, metruyencv.Host):
		c = metruyencv.New(cfg)
	case strings.Contains(host, tangthuvien.Host):
		c = tangthuvien.New(cfg)
	case strings.Contains(host, truyenchuHost):
		c = truyenchu.New(cfg)
	default:
		return nil, errors.New("")
	}

	return c, nil
}
