package crawlers

import (
	"errors"
	"strings"
	"sync"

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

func GetCrawler(host string, paths []string, ch *sync.Map) (Crawler, error) {
	var c Crawler

	switch {
	case strings.Contains(host, metruyencvHost):
		c = metruyencv.New(paths, ch)
	case strings.Contains(host, tangthuvienHost):
		c = tangthuvien.New(paths, ch)
	case strings.Contains(host, truyenchuHost):
		c = truyenchu.New(paths, ch)
	default:
		return nil, errors.New("")
	}

	return c, nil
}
