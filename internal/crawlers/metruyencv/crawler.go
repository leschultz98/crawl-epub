package metruyencv

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/epub"

	"github.com/PuerkitoBio/goquery"
)

const (
	Host                   = "metruyencv.com"
	host                   = "https://metruyencv.com/truyen"
	latestSelector         = "main button.rounded span.bg-primary"
	titleSelector          = "h2.text-center"
	contentSelector        = "#chapter-detail > div"
	contentLengthThreshold = 3000
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
		go func(i int) {
			defer wg.Done()

			url := fmt.Sprintf("%s/%s/chuong-%d", host, c.title, c.start+i)
			chapter, err := getChapter(url, 0)
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

func getChapter(origin string, mode int) (*epub.Chapter, error) {
	url := origin

	switch mode % 3 {
	case 1:
		url = strings.Replace(origin, host, "https://metruyencv.info/truyen", 1)
	case 2:
		url = strings.Replace(origin, host, "https://metruyencv-info.translate.goog/truyen", 1) + "?_x_tr_sl=auto&_x_tr_tl=en&_x_tr_hl=en&_x_tr_pto=wapp"
	}

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

	doc.Find(titleSelector).First().Each(func(_ int, s *goquery.Selection) {
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

	if (len(chapter.Content) < contentLengthThreshold) && mode < 12 {
		return getChapter(origin, mode+1)
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
	doc.Find(latestSelector).First().Each(func(i int, s *goquery.Selection) {
		latest, err = strconv.Atoi(s.Text())
	})
	if err != nil {
		return 0, err
	}

	return latest, nil
}

func makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return makeRequest(url)
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
