package tangthuvien

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/epub"

	"github.com/PuerkitoBio/goquery"
)

const (
	host            = "https://truyen-tangthuvien-vn.translate.goog/doc-truyen"
	idSelector      = "a.back"
	listSelector    = "li a[title]"
	titleSelector   = "h4.page-title"
	contentSelector = "p.content-block"
	suffix          = "_x_tr_sl=auto&_x_tr_tl=en&_x_tr_hl=en&_x_tr_pto=wapp"
)

type Crawler struct {
	title     string
	startPath string
	*config.Config
}

func New(c *config.Config) *Crawler {
	return &Crawler{
		title:     c.Paths[1],
		startPath: c.Paths[2],
		Config:    c,
	}
}

func (c *Crawler) GetEbook() (string, []*epub.Chapter, error) {
	id, err := getID(c.title, c.startPath)
	if err != nil {
		return "", nil, err
	}

	list, err := getList(id, c.startPath)
	if err != nil {
		c.Config.Info(err.Error())
		return "", nil, err
	}

	var wg sync.WaitGroup
	length := len(list)

	if c.MaxLength > 0 && length > c.MaxLength {
		list = list[0:c.MaxLength]
		length = c.MaxLength
	}

	chapters := make([]*epub.Chapter, length)
	wg.Add(length)

	for i := range list {
		go func(i int) {
			defer wg.Done()

			chapter, err := getChapter(list[i])
			if err != nil {
				panic(err)
			}

			chapters[i] = chapter
			c.Config.Info(chapter.Title)
			c.Config.Progress(length)
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

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return getChapter(url)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	chapter := &epub.Chapter{}

	doc.Find(titleSelector).Each(func(j int, s *goquery.Selection) {
		chapter.Title = s.Text()
	})

	chapter.Content = fmt.Sprintf("<h1>%s</h1>", chapter.Title)

	doc.Find(contentSelector).Each(func(j int, s *goquery.Selection) {
		chapter.Content += fmt.Sprintf("<p>%s</p>", s.Text())
	})

	return chapter, nil
}

func getID(title, startPath string) (string, error) {
	res, err := makeRequest(fmt.Sprintf("%s/%s/%s?%s", host, title, startPath, suffix))
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
		url := s.AttrOr("href", "")
		urlParts := strings.Split(url, "/")
		lastPart := urlParts[len(urlParts)-1]
		id = strings.Split(lastPart, "?")[0]
	})

	return id, nil
}

func getList(id string, startPath string) ([]string, error) {
	res, err := makeRequest(fmt.Sprintf("%s/page/%s?limit=9999&%s", host, id, suffix))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0)
	isStarted := startPath == ""

	doc.Find(listSelector).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr("href", "")

		if !isStarted && strings.Contains(url, startPath) {
			isStarted = true
		}

		if !isStarted {
			return
		}

		list = append(list, url)
	})

	return list, nil
}

func makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mobile")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
