package config

type Config struct {
	Paths      []string
	ProgressCh chan int
	InfoCh     chan string
	MaxLength  int
}

func (c *Config) Progress(num int) {
	if c.ProgressCh != nil {
		c.ProgressCh <- num
	}
}

func (c *Config) Info(text string) {
	if c.InfoCh != nil {
		c.InfoCh <- text
	}
}
