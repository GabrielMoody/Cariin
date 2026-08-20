package search

import (
	"encoding/json"
	"net/http"

	"github.com/GabrielMoody/Cariin/internal/documents"
	"github.com/GabrielMoody/Cariin/internal/index"
)

type SearchQuery struct {
	Q string `json:"query"`
}

type SearchResponse struct {
	Documents []documents.Document
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var q SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("error")
	}

	idx := index.Get()

	var docs []documents.Document

	for _, v := range (*idx)[q.Q] {
		docs = append(docs, documents.Documents[v])
	}

	data, _ := json.Marshal(SearchResponse{
		Documents: docs,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
