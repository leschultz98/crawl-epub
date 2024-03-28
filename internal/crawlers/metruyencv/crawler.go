package metruyencv

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/epub"

	"github.com/PuerkitoBio/goquery"
)

const (
	host            = "https://metruyencv.info/truyen"
	latestSelector  = "td a"
	titleSelector   = ".nh-read__title"
	contentSelector = "#article"
)

var ErrInvalidChapter = errors.New("invalid chapter")

type Crawler struct {
	title string
	start int
	*config.Config
}

func New(c *config.Config) *Crawler {
	return &Crawler{
		title:  c.Paths[1],
		start:  parseNumber(c.Paths[2]),
		Config: c,
	}
}

func (c *Crawler) GetEbook() (string, []*epub.Chapter, error) {
	latest, err := getLatest(c.title)
	if err != nil {
		return "", nil, err
	}

	var wg sync.WaitGroup
	length := latest - c.start + 1

	if c.MaxLength > 0 && length > c.MaxLength {
		length = c.MaxLength
	}

	chapters := make([]*epub.Chapter, length)
	end := length
	wg.Add(length)

	for i := 0; i < length; i++ {
		if i%70 == 0 {
			time.Sleep(200 * time.Millisecond)
		}

		go func(i int) {
			defer func() {
				wg.Done()
			}()

			url := fmt.Sprintf("%s/%s/chuong-%d", host, c.title, c.start+i)
			chapter, err := getChapter(url)
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
			c.Config.Info(chapter.Title)
			c.Config.Progress(length)
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

func getLatest(title string) (int, error) {
	res, err := makeRequest(fmt.Sprintf("%s/%s", host, title))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return 0, err
	}

	var latest int
	doc.Find(latestSelector).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr("href", "")
		urlParts := strings.Split(url, "/")
		latest = parseNumber(urlParts[len(urlParts)-1])
	})

	return latest, nil
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
