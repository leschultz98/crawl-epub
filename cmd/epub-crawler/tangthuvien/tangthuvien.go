package tangthuvien

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"crawl-epub/internal/epub"
	"crawl-epub/internal/progress"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

const idSelector = "a.back"
const listSelector = "li a[title]"
const contentSelector = "p.content-block"

type chapter struct {
	epub.Chapter
	url string
}

type Crawler struct {
	title     string
	startPath string
	bar       *progressbar.ProgressBar
}

func New(paths []string) *Crawler {
	return &Crawler{
		title:     paths[1],
		startPath: paths[2],
	}
}

func (c *Crawler) GetEbook() (string, []*epub.Chapter, error) {
	id, err := getID(c.title, c.startPath)
	list, err := getList(id, c.startPath)
	if err != nil {
		return "", nil, err
	}

	length := len(list)
	c.bar = progress.NewBar(length, "Get chapters...")
	chapters := make([]*epub.Chapter, 0, length)

	for i := range list {
		writeChapter(list[i])
		chapters = append(chapters, &list[i].Chapter)
		c.bar.Add(1)
	}

	return c.title, chapters, nil
}

func writeChapter(chapter *chapter) {
	res, err := makeRequest(chapter.url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		writeChapter(chapter)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	chapter.Content = fmt.Sprintf("<h1>%s</h1>", chapter.Title)

	doc.Find(contentSelector).Each(func(j int, s *goquery.Selection) {
		chapter.Content += fmt.Sprintf("<p>%s</p>", s.Text())
	})
}

func getID(title, startPath string) (string, error) {
	res, err := makeRequest(fmt.Sprintf("https://truyen.tangthuvien.vn/doc-truyen/%s/%s", title, startPath))
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

func getList(id string, startPath string) ([]*chapter, error) {
	res, err := makeRequest(fmt.Sprintf("https://truyen.tangthuvien.vn/doc-truyen/page/%s?limit=9999", id))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	chapters := make([]*chapter, 0)
	isStarted := startPath == ""

	doc.Find(listSelector).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr("href", "")

		if !isStarted && strings.Contains(url, startPath) {
			isStarted = true
		}

		if !isStarted {
			return
		}

		title := s.AttrOr("title", "")
		chapters = append(chapters, &chapter{
			Chapter: epub.Chapter{Title: title},
			url:     url,
		})
	})

	return chapters, nil
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
