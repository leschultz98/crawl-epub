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
	metruyencvHost  = "metruyencv.com"
	tangthuvienHost = "tangthuvien.vn"
	truyenchuHost   = "truyenchu.vn"
)

type Crawler interface {
	GetEbook() (string, []*epub.Chapter, error)
}

func GetCrawler(host string, cfg *config.Config) (Crawler, error) {
	var c Crawler

	switch {
	case strings.Contains(host, metruyencvHost):
		c = metruyencv.New(cfg)
	case strings.Contains(host, tangthuvienHost):
		c = tangthuvien.New(cfg)
	case strings.Contains(host, truyenchuHost):
		c = truyenchu.New(cfg)
	default:
		return nil, errors.New("")
	}

	return c, nil
}
