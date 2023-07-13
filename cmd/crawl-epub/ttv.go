package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const ttvListSelector = ".chapters:nth-child(2) li a"
const ttvContentSelector = "p.content-block"

type ttv struct{}

func (t ttv) getChapters(cfg *config) ([]*chapter, error) {
	client := &http.Client{}

	chapters, err := t.getChapterList(client, cfg)
	if err != nil {
		return nil, err
	}

	bar := newBar(len(chapters), "  Get chapter...")

	for i := range chapters {
		chapters[i].content = fmt.Sprintf("<h1>%s</h1>", chapters[i].title)
		res, err := t.makeRequest(client, chapters[i].url)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()

		if res.StatusCode != 200 {
			return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, err
		}

		doc.Find(ttvContentSelector).Each(func(j int, s *goquery.Selection) {
			chapters[i].content += fmt.Sprintf("<p>%s</p>", s.Text())
		})

		bar.Add(1)
	}

	return chapters, nil
}

func (t ttv) getChapterList(client *http.Client, cfg *config) ([]*chapter, error) {
	bar := newSpinner("Get chapter list...")
	defer bar.Finish()

	res, err := t.makeRequest(client, cfg.chapterListUrl)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	list := doc.Find(ttvListSelector)

	if cfg.length < 1 {
		cfg.length = list.Size()
	}

	chapters := make([]*chapter, 0, cfg.length)

	r, err := regexp.Compile(`(Chương\s+)(\d+)`)
	if err != nil {
		return nil, err
	}

	list.EachWithBreak(func(i int, s *goquery.Selection) bool {
		title := s.Text()
		url := s.AttrOr("href", "")

		chapters = append(chapters, &chapter{title: title, url: url})

		var number int
		number, err = strconv.Atoi(r.FindStringSubmatch(title)[2])
		if err != nil {
			return false
		}

		return number <= cfg.end
	})
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

func (t ttv) makeRequest(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mobile")

	return client.Do(req)
}
