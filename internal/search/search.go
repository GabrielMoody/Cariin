package search

import (
	"encoding/json"
	"net/http"
	"strings"

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
	var req SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("error")
	}

	idx := index.Get()
	queries := strings.Split(req.Q, " ")

	var docs []documents.Document

	for _, q := range queries {
		for _, v := range (*idx)[q] {
			docs = append(docs, documents.Documents[v])
		}
	}

	data, _ := json.Marshal(SearchResponse{
		Documents: docs,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
