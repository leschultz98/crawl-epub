package truyencv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/epub"

	"github.com/PuerkitoBio/goquery"
)

const host = "http://103.82.27.230:3001"

type chapterJSON struct {
	BuildId string `json:"buildId"`
	Props   struct {
		PageProps struct {
			Book struct {
				ID int `json:"id"`
			} `json:"book"`
			Chapter struct {
				Title       string `json:"title"`
				Content     string `json:"content"`
				NextChapter any    `json:"next_chapter"`
			} `json:"chapter"`
		} `json:"pageProps"`
	} `json:"props"`
}

type dataJSON struct {
	Data struct {
		Total int `json:"total"`
	} `json:"data"`
}

type Crawler struct {
	title     string
	startPath string
	length    int
	*config.Config
}

func New(c *config.Config) *Crawler {
	return &Crawler{
		title:     c.Paths[0],
		startPath: c.Paths[1],
		Config:    c,
	}
}

func (c *Crawler) GetEbook() (string, []*epub.Chapter, error) {
	var chapters []*epub.Chapter
	count := 1
	url := fmt.Sprintf("%s/%s/%s", host, c.title, c.startPath)

	for url != "" && count <= c.MaxLength {
		c.Config.Info(url)
		chapter, next, err := c.getChapter(url)
		if err != nil {
			panic(err)
		}

		chapters = append(chapters, chapter)
		c.Config.Progress(c.length)

		if next != "" {
			url = fmt.Sprintf("%s/%s/%s", host, c.title, next)
		} else {
			url = ""
		}
		count++
	}

	return c.title, chapters, nil
}

func (c *Crawler) getChapter(url string) (*epub.Chapter, string, error) {
	res, err := makeRequest(url)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, "", err
	}

	result := &epub.Chapter{}
	var next string

	doc.Find("#__NEXT_DATA__").Contents().Each(func(i int, s *goquery.Selection) {
		j := []byte(s.Text())

		parsedChapter := &chapterJSON{}
		err = json.Unmarshal(j, parsedChapter)
		if err != nil {
			return
		}

		//log.Println(parsedChapter.BuildId)
		if c.length == 0 {
			var total int
			total, err = getTotal(parsedChapter.Props.PageProps.Book.ID)
			if err != nil {
				return
			}

			var start int
			start, err = strconv.Atoi(strings.Split(c.startPath, "-")[1])
			if err != nil {
				return
			}

			c.length = total - start + 1
		}

		chapter := parsedChapter.Props.PageProps.Chapter
		chapter.Content = strings.ReplaceAll(chapter.Content, "p>", "section>")

		result.Title = chapter.Title

		if reflect.TypeOf(chapter.NextChapter).Name() != "bool" {
			next = chapter.NextChapter.(map[string]any)["slug"].(string)
		}

		var contentDoc *goquery.Document
		contentDoc, err = goquery.NewDocumentFromReader(strings.NewReader(chapter.Content))
		if err != nil {
			return
		}

		contentDoc.Find("section").Contents().Each(func(i int, s *goquery.Selection) {
			if goquery.NodeName(s) == "#text" && strings.TrimSpace(s.Text()) != "" {
				result.Content += fmt.Sprintf("<p>%s</p>", s.Text())
			}
		})
	})

	if err != nil {
		return nil, "", err
	}

	return result, next, nil
}

func getTotal(id int) (int, error) {
	res, err := makeRequest(fmt.Sprintf("%s/api/book?id=%d", host, id))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	d := &dataJSON{}
	err = json.NewDecoder(res.Body).Decode(d)
	if err != nil {
		return 0, err
	}

	return d.Data.Total, nil
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
