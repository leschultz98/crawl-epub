package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const (
	truyenchuTitleSelector       = ".chapter-text"
	truyenchuContentSelector     = "#chapter-c"
	truyenchuNextChapterSelector = "a#next_chap"
	truyenchuHost                = "https://truyenchu.vn"
)

type truyenchu struct{}

func (t *truyenchu) getChapters(cfg *config) ([]*chapter, error) {
	chapters := make([]*chapter, 0)

	url := fmt.Sprintf("%s/%s/%s", truyenchuHost, cfg.title, cfg.startURL)
	for url != "" {
		chapter, nextChapterURL, err := t.getChapter(url)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(chapter.title)

		url = nextChapterURL
		chapters = append(chapters, chapter)
	}

	return chapters, nil
}

func (t *truyenchu) getChapter(url string) (*chapter, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, "", err
	}

	var nextChapterURL string
	var title string
	var content string

	doc.Find(truyenchuNextChapterSelector).Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href != "#" {
			nextChapterURL = truyenchuHost + href
		}
	})

	doc.Find(truyenchuTitleSelector).Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			title = s.Text()
			content += fmt.Sprintf("<h1>%s</h1>", title)
		}
	})

	contentDoc := doc.Find(truyenchuContentSelector)

	contentDoc.Find("p").First().Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			content += fmt.Sprintf("<p>%s</p>", s.Text())
		}
	})

	contentDoc.Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			content += fmt.Sprintf("<p>%s</p>", s.Text())
		}
	})

	return &chapter{
		title:   title,
		content: content,
	}, nextChapterURL, nil
}
