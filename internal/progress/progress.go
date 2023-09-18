package progress

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

func NewBar(max int, desc string) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionShowCount(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Print("\n")
		}))
}
