package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const selector = "#inner_chap_content_1"

type chapter struct {
	title   string
	content string
}

func getChapter(url string) (*chapter, error) {
	log.Printf("get chapter from %s", url)

	res, err := http.Get(url)
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

	var title string
	var content string
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		titleEl := s.Find("p:first-child")
		titleEl.AddClass("title")
		title = titleEl.Text()
		content, err = s.Html()
	})
	if err != nil {
		return nil, err
	}

	return &chapter{
		title:   title,
		content: strings.TrimSpace(content),
	}, nil
}
