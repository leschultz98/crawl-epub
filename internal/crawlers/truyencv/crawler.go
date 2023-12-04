package truyencv

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/epub"

	"github.com/PuerkitoBio/goquery"
)

const host = "http://103.82.27.230:3001"

type ChapterData struct {
	PageProps struct {
		PageData struct {
			Book struct {
				TotalChapter int `json:"total_chapter"`
			} `json:"book"`
			Chapters []struct {
				Slug  string `json:"slug"`
				Title string `json:"title"`
			} `json:"chapters"`
		} `json:"pageData"`
	} `json:"pageProps"`
}

type Crawler struct {
	title        string
	startPath    string
	prefixURL    string
	startPage    int
	endPage      int
	startChapter int
	endChapter   int
	length       int
	chapterSlugs []string
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
	res, err := makeRequest(fmt.Sprintf("%s/%s", host, c.title))
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", nil, err
	}

	doc.Find("#__NEXT_DATA__").Contents().Each(func(i int, s *goquery.Selection) {
		j := []byte(s.Text())

		book := &struct {
			BuildId string      `json:"buildId"`
			Props   ChapterData `json:"props"`
		}{}

		err = json.Unmarshal(j, book)
		if err != nil {
			return
		}

		c.prefixURL = fmt.Sprintf("/_next/data/%s", book.BuildId)

		err = c.parseData(&book.Props, 1)
		if err != nil {
			return
		}
	})

	if err != nil {
		return "", nil, err
	}

	var outErr error
	var wg sync.WaitGroup
	wg.Add(c.endPage - c.startPage + 1)

	for i := c.startPage; i <= c.endPage; i++ {
		go func(i int) {
			defer func() {
				wg.Done()
			}()

			if i != 1 {
				res, err := makeRequest(fmt.Sprintf("%s/%s/%s.json?page=%d&book=%s", host, c.prefixURL, c.title, i, c.title))
				if err != nil {
					outErr = err
					return
				}
				defer res.Body.Close()

				d := &ChapterData{}
				err = json.NewDecoder(res.Body).Decode(d)
				if err != nil {
					outErr = err
					return
				}

				err = c.parseData(d, i)
				if err != nil {
					outErr = err
					return
				}
			}

		}(i)
	}
	wg.Wait()

	if outErr != nil {
		return "", nil, outErr
	}

	chapters := make([]*epub.Chapter, c.length)
	wg.Add(c.length)
	for i := range c.chapterSlugs {
		go func(i int) {
			defer func() {
				wg.Done()
			}()

			chapter, err := c.getChapter(c.chapterSlugs[i])
			if err != nil {
				outErr = err
				return
			}

			chapters[i] = chapter
			c.Config.Info(chapter.Title)
			c.Config.Progress(c.length)
		}(i)
	}
	wg.Wait()

	if outErr != nil {
		return "", nil, outErr
	}

	return c.title, chapters, nil
}

func (c *Crawler) parseData(data *ChapterData, page int) error {
	c.Config.Info(fmt.Sprintf("%s page %d...", c.title, page))
	if c.length == 0 {
		start, err := strconv.Atoi(strings.Split(c.startPath, "-")[1])
		if err != nil {
			return err
		}

		c.startChapter = start
		c.length = data.PageProps.PageData.Book.TotalChapter - c.startChapter + 1

		if c.length > c.MaxLength {
			c.length = c.MaxLength
		}

		c.endChapter = c.startChapter + c.length - 1

		c.chapterSlugs = make([]string, c.length)

		c.endPage = int(math.Ceil(float64(c.endChapter) / 50))
		c.startPage = int(math.Ceil(float64(c.startChapter) / 50))

		if c.startPage > page {
			return nil
		}
	}

	for i, chapter := range data.PageProps.PageData.Chapters {
		index := (page-1)*50 + i + 1 - c.startChapter
		if index < 0 {
			continue
		}

		if index == c.length {
			break
		}

		c.chapterSlugs[index] = chapter.Slug
	}

	return nil
}

func (c *Crawler) getChapter(slug string) (*epub.Chapter, error) {
	res, err := makeRequest(fmt.Sprintf("%s/%s/%s/%s.json?book=%s&chapter=%s", host, c.prefixURL, c.title, slug, c.title, slug))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		c.Config.Info("Retry " + slug)
		return c.getChapter(slug)
	}

	d := &struct {
		PageProps struct {
			Chapter struct {
				Content string `json:"content"`
				Title   string `json:"title"`
			} `json:"chapter"`
		} `json:"pageProps"`
	}{}

	err = json.NewDecoder(res.Body).Decode(d)
	if err != nil {
		return nil, err
	}

	result := &epub.Chapter{}
	result.Title = d.PageProps.Chapter.Title
	result.Content += fmt.Sprintf("<h1>%s</h1>", result.Title)

	content := strings.ReplaceAll(d.PageProps.Chapter.Content, "p>", "section>")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	doc.Find("section").Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" && strings.TrimSpace(s.Text()) != "" {
			result.Content += fmt.Sprintf("<p>%s</p>", s.Text())
		}
	})

	return result, nil
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
