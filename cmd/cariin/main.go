package main

import (
	"log"
	"net/http"

	"github.com/GabrielMoody/Cariin/internal/index"
	"github.com/GabrielMoody/Cariin/internal/search"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("POST /search", search.SearchHandler)

	log.Println("Building inverted index")

	index.Init()

	log.Println("Inverted index build")

	log.Println("Server running on port 8000")

	http.ListenAndServe("localhost:8000", router)
}
