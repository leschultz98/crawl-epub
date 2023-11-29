package epub

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/bmaupin/go-epub"
)

const (
	fontPath  = "assets/fonts/OpenSans-Regular.ttf"
	cssPath   = "assets/styles/styles.css"
	outputDir = "ebooks"
	MaxS      = 300
)

type Chapter struct {
	Title   string
	Content string
}

func WriteTo(w io.Writer, title string, chapters []*Chapter) error {
	e, err := createEpub(title, chapters)
	if err != nil {
		return err
	}

	_, err = e.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}

func WriteEpub(title string, chapters []*Chapter) error {
	length := len(chapters)
	count := int(math.Ceil(float64(length) / float64(MaxS)))

	for i := 0; i < count; i++ {
		suffix := fmt.Sprintf("-%d", i+1)
		if count == 1 {
			suffix = ""
		}

		end := MaxS * (i + 1)
		if end > length {
			end = length
		}

		e, err := createEpub(title+suffix, chapters[i*(MaxS):end])
		if err != nil {
			return err
		}

		err = writeLocal(title+suffix, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func createEpub(title string, chapters []*Chapter) (*epub.Epub, error) {
	e := epub.NewEpub(title)

	_, err := e.AddFont(fontPath, "")
	if err != nil {
		return nil, err
	}

	css, err := e.AddCSS(cssPath, "")
	if err != nil {
		return nil, err
	}

	for _, chapter := range chapters {
		if chapter.Title != "" && chapter.Content != "" {
			_, err := e.AddSection(chapter.Content, chapter.Title, "", css)
			if err != nil {
				return nil, err
			}
		}
	}

	return e, nil
}

func writeLocal(title string, e *epub.Epub) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
			return err
		}
	}

	err := e.Write(fmt.Sprintf("%s/%s.epub", outputDir, title))
	if err != nil {
		return err
	}

	return nil
}
