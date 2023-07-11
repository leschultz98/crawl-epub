package main

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

func newBar(max int, desc string) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Print("\n")
		}))
}
