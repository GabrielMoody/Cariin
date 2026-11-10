package main

import (
	"log"
	"net/http"

	"github.com/GabrielMoody/Cariin/internal/index"
	"github.com/GabrielMoody/Cariin/internal/search"
	"github.com/GabrielMoody/Cariin/internal/web"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("POST /search", search.SearchHandler)
	router.Handle("GET /", web.Handler())

	log.Println("Building inverted index")

	index.Init()

	log.Println("Inverted index build")

	log.Println("Server running on port 8000")

	if err := http.ListenAndServe("localhost:8000", router); err != nil {
		log.Fatal(err)
	}
}
