package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const titleSelector = "h2.heading-font"
const contentSelector = "#inner_chap_content_1"

type chapter struct {
	title   string
	content string
}

func (app *application) getChapter(url string, number int) (*chapter, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header["User-Agent"] = []string{"undici"}

	res, err := app.client.Do(req)
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

	doc.Find(titleSelector).Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})

	doc.Find(contentSelector).Each(func(i int, s *goquery.Selection) {
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
