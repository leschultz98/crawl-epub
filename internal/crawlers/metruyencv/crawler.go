package metruyencv

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"crawl-epub/internal/epub"
	"crawl-epub/internal/progress"

	"github.com/PuerkitoBio/goquery"
)

const (
	host            = "https://metruyencv.com/truyen"
	maxSelector     = "td a"
	titleSelector   = ".nh-read__title"
	contentSelector = "#article"
)

var ErrInvalidChapter = errors.New("invalid chapter")

type Crawler struct {
	title string
	start int
	ch    *sync.Map
}

func New(paths []string, ch *sync.Map) *Crawler {
	return &Crawler{
		title: paths[1],
		start: parseNumber(paths[2]),
		ch:    ch,
	}
}

func (c *Crawler) GetEbook(maxLength int) (string, []*epub.Chapter, error) {
	max, err := getMax(c.title)
	if err != nil {
		return "", nil, err
	}

	var wg sync.WaitGroup
	length := max - c.start + 1

	if maxLength > 0 && length > maxLength {
		length = maxLength
	}

	chapters := make([]*epub.Chapter, length)
	end := length
	bar := progress.NewBar(length, "Get chapters...")
	wg.Add(length)

	for i := 0; i < length; i++ {
		if i%70 == 0 {
			time.Sleep(200 * time.Millisecond)
		}

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

			chapter, err := getChapter(fmt.Sprintf("%s/%s/chuong-%d", host, c.title, c.start+i))
			if err != nil {
				if errors.Is(err, ErrInvalidChapter) {
					if end > i {
						end = i
					}
					return
				}

				panic(err)
			}

			chapters[i] = chapter

		}(i)
	}

	wg.Wait()
	return c.title, chapters[:end], nil
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

	doc.Find(titleSelector).Each(func(_ int, s *goquery.Selection) {
		chapter.Title = strings.TrimSpace(s.Text())
		chapter.Content = fmt.Sprintf("<h1>%s</h1>", chapter.Title)
	})

	doc.Find(contentSelector).Contents().EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if goquery.NodeName(s) == "div" && s.Text() == "Vui lòng đăng nhập để đọc tiếp nội dung" {
			err = ErrInvalidChapter
			return false
		}

		if goquery.NodeName(s) == "#text" {
			chapter.Content += fmt.Sprintf("<p>%s</p>", strings.TrimSpace(s.Text()))
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	return chapter, nil
}

func getMax(title string) (int, error) {
	res, err := makeRequest(fmt.Sprintf("%s/%s", host, title))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return 0, err
	}

	var max int
	doc.Find(maxSelector).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr("href", "")
		urlParts := strings.Split(url, "/")
		max = parseNumber(urlParts[len(urlParts)-1])
	})

	return max, nil
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

func parseNumber(path string) int {
	num, err := strconv.Atoi(strings.Split(path, "-")[1])
	if err != nil {
		panic(err)
	}

	return num
}
