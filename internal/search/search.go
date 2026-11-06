package search

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/GabrielMoody/Cariin/internal/documents"
	"github.com/GabrielMoody/Cariin/internal/index"
	"github.com/GabrielMoody/Cariin/internal/reranker"
	"github.com/GabrielMoody/Cariin/internal/search/parser"
	"github.com/GabrielMoody/Cariin/internal/search/ranking"
)

type SearchQuery struct {
	Q string `json:"query"`
}

type SearchResponse struct {
	Documents []documents.Document
}

var (
	client *reranker.Client
	once   sync.Once
)

func GetHttpClient() *reranker.Client {
	once.Do(func() {
		client = reranker.NewClient("http://127.0.0.1:8001")
	})

	return client
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var req SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	query, err := parser.Parse(req.Q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idx := index.Get()
	postings := query.Evaluate(*idx)

	bm := ranking.NewBM25()

	postings = bm.RankBM25(query, *idx, postings)
	docs := make([]documents.Document, 0, len(postings))
	for _, posting := range postings {
		docs = append(docs, documents.Documents[posting.DocId])
	}

	log.Printf("Retrieve %d docs\n", len(docs))

	client := GetHttpClient()

	res, err := client.Rerank(r.Context(), req.Q, docs)

	if err != nil {
		log.Println(err.Error())
	}

	data, _ := json.Marshal(reranker.RerankResponse{
		Query:   req.Q,
		Results: res.Results,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
