package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const url = "http://103.82.27.230:3001/tien-tu-xin-nghe-ta-giai-thich/chuong-458-tran-phap"

type chapterJSON struct {
	Props struct {
		PageProps struct {
			Chapter struct {
				Title       string `json:"title"`
				Content     string `json:"content"`
				NextChapter any    `json:"next_chapter"`
				PrevChapter any    `json:"prev_chapter"`
			} `json:"chapter"`
		} `json:"pageProps"`
	} `json:"props"`
}

func main() {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#__NEXT_DATA__").Contents().Each(func(i int, s *goquery.Selection) {
		j := []byte(s.Text())

		err = os.WriteFile("cmd/test/j.json", j, 0644)
		if err != nil {
			log.Fatal(err)
		}

		c := &chapterJSON{}
		err := json.Unmarshal(j, c)
		if err != nil {
			log.Fatal(err)
		}

		chapter := c.Props.PageProps.Chapter

		fmt.Println(chapter.Title)
		fmt.Println(chapter.NextChapter)
		fmt.Println(chapter.PrevChapter)
		fmt.Println(reflect.TypeOf(chapter.NextChapter).Name() == "bool")
		fmt.Println(reflect.TypeOf(chapter.PrevChapter.(map[string]any)["slug"]))

		chapter.Content = strings.ReplaceAll(chapter.Content, "p>", "section>")

		err = os.WriteFile("cmd/test/content.html", []byte(chapter.Content), 0644)
		if err != nil {
			log.Fatal(err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(chapter.Content))
		if err != nil {
			log.Fatal(err)
		}

		var content string

		doc.Find("section").Contents().Each(func(i int, s *goquery.Selection) {
			if goquery.NodeName(s) == "#text" && strings.TrimSpace(s.Text()) != "" {
				content += fmt.Sprintf("<p>%s</p>", s.Text())
			}
		})

		err = os.WriteFile("cmd/test/parsed.html", []byte(content), 0644)
		if err != nil {
			log.Fatal(err)
		}
	})
}
