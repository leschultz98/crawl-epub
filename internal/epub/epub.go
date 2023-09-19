package epub

import (
	"fmt"
	"math"
	"os"

	"github.com/bmaupin/go-epub"
)

const (
	fontPath  = "assets/fonts/OpenSans-Regular.ttf"
	cssPath   = "assets/styles/styles.css"
	outputDir = "ebooks"
	maxS      = 300
)

type Chapter struct {
	Title   string
	Content string
}

func WriteEpub(title string, chapters []*Chapter) error {
	length := len(chapters)
	count := int(math.Ceil(float64(length) / float64(maxS)))

	for i := 0; i < count; i++ {
		suffix := fmt.Sprintf("-%d", i+1)
		if count == 1 {
			suffix = ""
		}

		end := maxS * (i + 1)
		if end > length {
			end = length
		}

		err := write(title+suffix, chapters[i*(maxS):end])
		if err != nil {
			return err
		}
	}

	return nil
}

func write(title string, chapters []*Chapter) error {
	e := epub.NewEpub(title)

	_, err := e.AddFont(fontPath, "")
	if err != nil {
		return err
	}

	css, err := e.AddCSS(cssPath, "")
	if err != nil {
		return err
	}

	for _, chapter := range chapters {
		if chapter.Title != "" && chapter.Content != "" {
			_, err := e.AddSection(chapter.Content, chapter.Title, "", css)
			if err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
			return err
		}
	}

	err = e.Write(fmt.Sprintf("%s/%s.epub", outputDir, title))
	if err != nil {
		return err
	}

	return nil
}
