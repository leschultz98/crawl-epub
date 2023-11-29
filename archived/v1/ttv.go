package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

const ttvListSelector = "li a[title]"
const ttvContentSelector = "p.content-block"

type ttv struct {
	errorLog *log.Logger
	client   *http.Client
	chapters []*chapter
	bar      *progressbar.ProgressBar
}

func (t *ttv) getChapters(cfg *config) ([]*chapter, error) {
	t.client = &http.Client{}

	chapters, err := t.getChapterList(cfg)
	if err != nil {
		return nil, err
	}

	length := len(chapters)
	t.chapters = chapters
	t.bar = newBar(length, "  Get chapter...")

	for i := range t.chapters {
		req, err := t.newRequest(chapters[i].url)
		if err != nil {
			return nil, err
		}

		t.getChapter(req, i)
	}

	return chapters, nil
}

func (t *ttv) getChapter(req *http.Request, i int) {
	res, err := t.client.Do(req)
	if err != nil {
		t.errorLog.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		t.getChapter(req, i)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.errorLog.Fatal(err)
	}

	t.chapters[i].content = fmt.Sprintf("<h1>%s</h1>", t.chapters[i].title)

	doc.Find(ttvContentSelector).Each(func(j int, s *goquery.Selection) {
		t.chapters[i].content += fmt.Sprintf("<p>%s</p>", s.Text())
	})

	t.bar.Add(1)
}

func (t *ttv) getChapterList(cfg *config) ([]*chapter, error) {
	bar := newSpinner("Get chapter list...")
	defer bar.Finish()

	req, err := t.newRequest(fmt.Sprintf("https://truyen.tangthuvien.vn/doc-truyen/page/%s?limit=10000", cfg.bookID))
	if err != nil {
		return nil, err
	}

	res, err := t.client.Do(req)
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
	listLength := list.Size()
	chapters := make([]*chapter, 0, listLength)
	isStarted := cfg.startURL == ""

	list.EachWithBreak(func(i int, s *goquery.Selection) bool {
		url := s.AttrOr("href", "")

		if !isStarted && strings.Contains(url, cfg.startURL) {
			isStarted = true
		}

		if !isStarted {
			return true
		}

		url = strings.Replace(url, "https://truyen.tangthuvien.vn/", "https://m.truyen.tangthuvien.vn/", 1)
		title := s.AttrOr("title", "")
		chapters = append(chapters, &chapter{title: title, url: url})

		return cfg.endURL == "" || !strings.Contains(url, cfg.endURL)
	})

	return chapters, nil
}

func (t *ttv) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mobile")

	return req, nil
}
