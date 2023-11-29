package crawlers

import (
	"errors"
	"strings"

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
	GetEbook(maxLength int) (string, []*epub.Chapter, error)
}

func GetCrawler(host string, paths []string) (Crawler, error) {
	var c Crawler

	switch {
	case strings.Contains(host, metruyencvHost):
		c = metruyencv.New(paths)
	case strings.Contains(host, tangthuvienHost):
		c = tangthuvien.New(paths)
	case strings.Contains(host, truyenchuHost):
		c = truyenchu.New(paths)
	default:
		return nil, errors.New("")
	}

	return c, nil
}
