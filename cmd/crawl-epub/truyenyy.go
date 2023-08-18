package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const (
	truyenyyTitleSelector   = "h2.heading-font"
	truyenyyContentSelector = "#inner_chap_content_1"
	truyenyyUrlFormat       = "https://truyenyy.vip/truyen/%s/chuong-%d.html"
)

type truyenyy struct {
	client *http.Client
	wg     sync.WaitGroup
}

func (t *truyenyy) getChapters(cfg *config) ([]*chapter, error) {
	if cfg.length < 1 {
		log.Fatal("must set appropriate end")
	}

	bar := newBar(cfg.length, "  Get chapter...")

	t.client = &http.Client{
		// disable HTTP/2
		Transport: &http.Transport{
			TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
		},
	}

	t.wg.Add(cfg.length)

	chapters := make([]*chapter, cfg.length)

	for i := 0; i < cfg.length; i++ {
		go func(i int) {
			chapter, err := t.getChapter(fmt.Sprintf(truyenyyUrlFormat, cfg.title, i+1), i+1)
			if err != nil {
				log.Fatal(err)
			}

			chapters[i] = chapter

			bar.Add(1)
			t.wg.Done()
		}(i)
	}

	t.wg.Wait()
	return chapters, nil
}

func (t *truyenyy) getChapter(url string, number int) (*chapter, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header["User-Agent"] = []string{"undici"}

	res, err := t.client.Do(req)
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

	doc.Find(truyenyyTitleSelector).Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})

	doc.Find(truyenyyContentSelector).Each(func(i int, s *goquery.Selection) {
		content, err = s.Html()
	})
	if err != nil {
		return nil, err
	}

	title = fmt.Sprintf("Chương %d: %s", number, strings.TrimSpace(title))
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, fmt.Sprintf("<p>%s</p>", title))
	content = strings.TrimSpace(content)
	content = fmt.Sprintf("<h1>%s</h1>\n%s", title, content)

	return &chapter{
		title:   title,
		content: content,
	}, nil
}
