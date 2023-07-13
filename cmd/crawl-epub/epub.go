package main

import (
	"fmt"
	"os"

	"github.com/bmaupin/go-epub"
)

const (
	fontPath  = "assets/fonts/OpenSans-Regular.ttf"
	cssPath   = "assets/styles/styles.css"
	outputDir = "ebooks"
)

func writeEpub(title string, chapters []*chapter) error {
	newSpinner("Write epub...")

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
