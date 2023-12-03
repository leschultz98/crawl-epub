package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"crawl-epub/internal/crawlers"
	"crawl-epub/internal/crawlers/config"
	"crawl-epub/internal/epub"
	"crawl-epub/public"
)

type app struct {
	chMap *sync.Map
}

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

	a := app{
		chMap: &sync.Map{},
	}

	mux.Handle("/", http.FileServer(http.FS(public.StaticFiles)))
	mux.HandleFunc("/messages", a.messagesHandler)
	mux.HandleFunc("/api/", a.crawlHandler)

	log.Printf("Started at: http://localhost:%s\n", port)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func (a *app) crawlHandler(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	host := urlParts[2]
	paths := urlParts[3:]

	id := r.URL.Query().Get("id")
	ch, _ := a.chMap.Load(id)

	cfg := &config.Config{
		Paths:     paths,
		Ch:        ch.(chan int),
		MaxLength: epub.MaxS,
	}

	c, err := crawlers.GetCrawler(host, cfg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	title, chapters, err := c.GetEbook()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+title+".epub")

	err = epub.WriteTo(w, title, chapters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *app) messagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := r.URL.Query().Get("id")

	c := make(chan int, epub.MaxS)
	a.chMap.Store(id, c)

	for {
		select {
		case num := <-c:
			fmt.Fprintf(w, "data: %d\n\n", num)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			a.chMap.Delete(id)
		}
	}
}
