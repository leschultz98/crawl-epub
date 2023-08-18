package main

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

func writeEpubs(title string, chapters []*chapter) error {
	newSpinner("Write epubs...")

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

		err := writeEpub(title+suffix, chapters[i*(maxS):end])
		if err != nil {
			return err
		}
	}

	return nil
}

func writeEpub(title string, chapters []*chapter) error {
	e := epub.NewEpub(title)

	_, err := e.AddFont(fontPath, "")
	if err != nil {
		return err
	}

	css, err := e.AddCSS(cssPath, "")
	if err != nil {
		return err
	}

	for i := range chapters {
		_, err := e.AddSection(chapters[i].content, chapters[i].title, "", css)
		if err != nil {
			return err
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
