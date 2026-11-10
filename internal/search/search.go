package search

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
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
	Documents []documents.Document `json:"documents"`
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
	if len(docs) == 0 {
		writeSearchResponse(w, SearchResponse{Documents: docs})
		return
	}

	client := GetHttpClient()

	res, err := client.Rerank(r.Context(), req.Q, docs)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "reranking failed", http.StatusBadGateway)
		return
	}

	writeSearchResponse(w, SearchResponse{
		Documents: documentsByRank(res.Results, docs),
	})
}

func writeSearchResponse(w http.ResponseWriter, response SearchResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "failed to encode search results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("write search response: %v", err)
	}
}

func documentsByRank(results []reranker.RankedDocument, docs []documents.Document) []documents.Document {
	rankedResults := append([]reranker.RankedDocument(nil), results...)
	sort.SliceStable(rankedResults, func(i, j int) bool {
		return rankedResults[i].Rank < rankedResults[j].Rank
	})

	documentsByID := make(map[int64]documents.Document, len(docs))
	for _, doc := range docs {
		documentsByID[doc.Id] = doc
	}

	rankedDocuments := make([]documents.Document, 0, len(rankedResults))
	for _, result := range rankedResults {
		if doc, ok := documentsByID[result.ID]; ok {
			rankedDocuments = append(rankedDocuments, doc)
		}
	}

	return rankedDocuments
}
