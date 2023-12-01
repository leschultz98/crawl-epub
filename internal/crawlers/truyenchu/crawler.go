package truyenchu

import (
	"fmt"
	"net/http"
	"sync"

	"crawl-epub/internal/epub"
	"crawl-epub/internal/progress"

	"github.com/PuerkitoBio/goquery"
)

const (
	host            = "https://truyenchu.vn"
	idSelector      = "#truyen-id"
	listSelector    = "option"
	titleSelector   = ".chapter-text"
	contentSelector = "#chapter-c"
)

type Crawler struct {
	title     string
	startPath string
	ch        *sync.Map
}

func New(paths []string, ch *sync.Map) *Crawler {
	return &Crawler{
		title:     paths[0],
		startPath: paths[1],
		ch:        ch,
	}
}

func (c *Crawler) GetEbook(maxLength int) (string, []*epub.Chapter, error) {
	id, err := getID(c.title, c.startPath)
	if err != nil {
		return "", nil, err
	}

	list, err := getList(id, c.title, c.startPath)
	if err != nil {
		return "", nil, err
	}

	var wg sync.WaitGroup
	length := len(list)

	if maxLength > 0 && length > maxLength {
		list = list[0:maxLength]
		length = maxLength
	}

	chapters := make([]*epub.Chapter, length)
	bar := progress.NewBar(length, "Get chapters...")
	wg.Add(length)

	for i := range list {
		go func(i int) {
			defer func() {
				bar.Add(1)
				if c.ch != nil {
					c.ch.Range(func(key, value any) bool {
						value.(chan int) <- length
						return true
					})
				}
				wg.Done()
			}()

			chapter, err := getChapter(list[i])
			if err != nil {
				panic(err)
			}

			chapters[i] = chapter
		}(i)
	}

	wg.Wait()
	return c.title, chapters, nil
}

func getChapter(url string) (*epub.Chapter, error) {
	res, err := makeRequest(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	chapter := &epub.Chapter{}

	doc.Find(titleSelector).Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			chapter.Title = s.Text()
			chapter.Content += fmt.Sprintf("<h1>%s</h1>", chapter.Title)
		}
	})

	contentDoc := doc.Find(contentSelector)

	contentDoc.Find("p").First().Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			chapter.Content += fmt.Sprintf("<p>%s</p>", s.Text())
		}
	})

	contentDoc.Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			chapter.Content += fmt.Sprintf("<p>%s</p>", s.Text())
		}
	})

	return chapter, nil
}

func getID(title, startPath string) (string, error) {
	res, err := makeRequest(fmt.Sprintf("%s/%s/%s", host, title, startPath))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	var id string
	doc.Find(idSelector).Each(func(i int, s *goquery.Selection) {
		id = s.AttrOr("value", "")
	})

	return id, nil
}

func getList(id, title, startPath string) ([]string, error) {
	res, err := makeRequest(fmt.Sprintf("%s/api/services/chapter-option?type=chapter_option&data=%s", host, id))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0)
	isStarted := false

	doc.Find(listSelector).Each(func(i int, s *goquery.Selection) {
		path := s.AttrOr("value", "")

		if !isStarted && startPath == path {
			isStarted = true
		}

		if !isStarted {
			return
		}

		list = append(list, fmt.Sprintf("%s/%s/%s", host, title, path))
	})

	return list, nil
}

func makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
