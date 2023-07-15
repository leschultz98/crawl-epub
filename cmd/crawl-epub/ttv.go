package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

const ttvListSelector = "li a"
const ttvContentSelector = "p.content-block"

type ttv struct {
	client   *http.Client
	chapters []*chapter
	bar      *progressbar.ProgressBar
	wg       sync.WaitGroup
	l        sync.Mutex
}

func (t *ttv) getChapters(cfg *config) ([]*chapter, error) {
	t.client = &http.Client{}

	chapters, err := t.getChapterList(cfg)
	if err != nil {
		return nil, err
	}

	t.chapters = chapters
	t.bar = newBar(cfg.length, "  Get chapter...")
	t.wg.Add(cfg.length)

	for i := range t.chapters {
		req, err := t.newRequest(chapters[i].url)
		if err != nil {
			return nil, err
		}

		go t.getChapter(req, i)
	}

	t.wg.Wait()
	return chapters, nil
}

func (t *ttv) getChapter(req *http.Request, i int) {
	res, err := t.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusServiceUnavailable {
		res.Body.Close()
		go t.getChapter(req, i)
		return
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal(fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	t.l.Lock()
	t.chapters[i].content = fmt.Sprintf("<h1>%s</h1>", t.chapters[i].title)

	doc.Find(ttvContentSelector).Each(func(j int, s *goquery.Selection) {
		t.chapters[i].content += fmt.Sprintf("<p>%s</p>", s.Text())
	})
	t.l.Unlock()

	t.bar.Add(1)
	t.wg.Done()
}

func (t *ttv) getChapterList(cfg *config) ([]*chapter, error) {
	bar := newSpinner("Get chapter list...")
	defer bar.Finish()

	req, err := t.newRequest(fmt.Sprintf("https://truyen.tangthuvien.vn/doc-truyen/page/%s?limit=10000", cfg.bookID))
	if err != nil {
		return nil, err
	}

	res, err := t.client.Do(req)
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

	list := doc.Find(ttvListSelector)

	if cfg.length < 1 {
		cfg.length = list.Size()
	}

	chapters := make([]*chapter, 0, cfg.length)

	r, err := regexp.Compile(`(Chương\s+)(\d+)`)
	if err != nil {
		return nil, err
	}

	list.EachWithBreak(func(i int, s *goquery.Selection) bool {
		title := s.AttrOr("title", "")

		var number int
		number, err = strconv.Atoi(r.FindStringSubmatch(title)[2])
		if err != nil {
			return false
		}

		if number > cfg.end {
			return false
		}

		if number >= cfg.from {
			url := s.AttrOr("href", "")
			url = strings.Replace(url, "https://truyen.tangthuvien.vn/", "https://m.truyen.tangthuvien.vn/", 1)
			chapters = append(chapters, &chapter{title: title, url: url})
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

func (t *ttv) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mobile")

	return req, nil
}
