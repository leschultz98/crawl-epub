package tangthuvien

import (
	"fmt"
	"net/http"
	"strings"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/epub"

	"github.com/PuerkitoBio/goquery"
)

const (
	host            = "https://truyen.tangthuvien.vn/doc-truyen"
	idSelector      = "a.back"
	listSelector    = "li a[title]"
	titleSelector   = "h4.page-title"
	contentSelector = "p.content-block"
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

	length := len(list)

	if c.MaxLength > 0 && length > c.MaxLength {
		list = list[0:c.MaxLength]
		length = c.MaxLength
	}

	chapters := make([]*epub.Chapter, 0, length)

	for i := range list {
		c.Config.Info(list[i])
		chapter, err := getChapter(list[i])
		if err != nil {
			return "", nil, err
		}

		chapters = append(chapters, chapter)
		c.Config.Progress(length)
	}

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
		url := s.AttrOr("href", "")
		urlParts := strings.Split(url, "/")
		id = urlParts[len(urlParts)-1]
	})

	return id, nil
}

func getList(id string, startPath string) ([]string, error) {
	res, err := makeRequest(fmt.Sprintf("%s/page/%s?limit=9999", host, id))
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
