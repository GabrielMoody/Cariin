package main

import (
	"log"
	"net/http"

	"github.com/GabrielMoody/Cariin/internal/crawler"
	"github.com/GabrielMoody/Cariin/internal/search"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("POST /search", search.SearchHandler)

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	log.Println("Start crawling websites")

	crawler.Crawl()

	log.Println("Server running on port 8000")

	http.ListenAndServe("localhost:8000", router)
}
