package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"crawl-epub/internal/crawlers"
	"crawl-epub/internal/epub"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("/", crawler)

	log.Fatal(s.ListenAndServe())
}

func crawler(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	host := urlParts[1]
	paths := urlParts[2:]

	c, err := crawlers.GetCrawler(host, paths)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	title, chapters, err := c.GetEbook(epub.MaxS)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+title+".epub")

	err = epub.WriteTo(w, title, chapters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
