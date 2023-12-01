package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"crawl-epub/internal/crawlers"
	"crawl-epub/internal/epub"
	"crawl-epub/public"

	"github.com/google/uuid"
)

type app struct {
	ch *sync.Map
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
		ch: &sync.Map{},
	}

	mux.Handle("/", http.FileServer(http.FS(public.StaticFiles)))
	mux.HandleFunc("/messages", a.messagesHandler)
	mux.HandleFunc("/api/", a.crawlHandler)

	log.Println("Started on port", port)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func (a *app) crawlHandler(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	host := urlParts[2]
	paths := urlParts[3:]

	c, err := crawlers.GetCrawler(host, paths, a.ch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	title, chapters, err := c.GetEbook(epub.MaxS)
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

	id := uuid.New().String()

	c := make(chan int, epub.MaxS)
	a.ch.Store(id, c)

	for {
		select {
		case num := <-c:
			fmt.Fprintf(w, "data: %d\n\n", num)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			a.ch.Delete(id)
		}
	}

}
