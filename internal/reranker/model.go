package reranker

type Document struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Text  string `json:"text"`
}

type RerankRequest struct {
	Query     string     `json:"query"`
	Documents []Document `json:"documents"`
}

type RankedDocument struct {
	ID    int64   `json:"id"`
	Score float64 `json:"score"`
	Rank  int     `json:"rank"`
}

type RerankResponse struct {
	Query   string           `json:"query"`
	Model   string           `json:"model"`
	Results []RankedDocument `json:"results"`
}
