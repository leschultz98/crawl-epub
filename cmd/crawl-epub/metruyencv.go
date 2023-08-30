package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	metruyencvTitleSelector   = ".nh-read__title"
	metruyencvContentSelector = "#article"
	metruyencvUrlFormat       = "https://metruyencv.com/truyen/%s/chuong-%d"
)

var ErrInvalidChapter = errors.New("invalid chapter")

type metruyencv struct {
	wg sync.WaitGroup
}

func (t *metruyencv) getChapters(cfg *config) ([]*chapter, error) {
	length := cfg.end - cfg.start + 1
	bar := newBar(length, "  Get chapter...")
	end := length
	chapters := make([]*chapter, length)

	t.wg.Add(length)
	for i := 0; i < length; i++ {
		if i%70 == 0 {
			time.Sleep(200 * time.Millisecond)
		}
		go func(i int) {
			defer func() {
				bar.Add(1)
				t.wg.Done()
			}()

			chapter, err := t.getChapter(fmt.Sprintf(metruyencvUrlFormat, cfg.title, cfg.start+i))
			if err != nil {
				if errors.Is(err, ErrInvalidChapter) {
					if end > i {
						end = i
					}
					return
				}

				log.Fatal(err)
			}

			chapters[i] = chapter

		}(i)
	}
	t.wg.Wait()

	return chapters[:end], nil
}

func (t *metruyencv) getChapter(url string) (*chapter, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var title string
	var content string

	doc.Find(metruyencvTitleSelector).Each(func(_ int, s *goquery.Selection) {
		title = strings.TrimSpace(s.Text())
		content = fmt.Sprintf("<h1>%s</h1>", title)
	})

	doc.Find(metruyencvContentSelector).Contents().EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if goquery.NodeName(s) == "div" && s.Text() == "Vui lòng đăng nhập để đọc tiếp nội dung" {
			err = ErrInvalidChapter
			return false
		}

		if goquery.NodeName(s) == "#text" {
			content += fmt.Sprintf("<p>%s</p>", strings.TrimSpace(s.Text()))
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	return &chapter{
		title:   title,
		content: content,
	}, nil
}
