package reranker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/GabrielMoody/Cariin/internal/documents"
)

const (
	MaxDocuments          = 5000
	MaxQueryLength        = 1000
	MaxDocumentTextLength = 8000
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 60 * time.Minute,
		},
	}
}

// NewClientWithHTTP allows custom timeout and transport configuration.
func NewClientWithHTTP(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) Rerank(ctx context.Context, query string, documents []documents.Document) (*RerankResponse, error) {

	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if len([]rune(query)) > MaxQueryLength {
		return nil, fmt.Errorf(
			"query exceeds maximum length of %d characters",
			MaxQueryLength,
		)
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("documents cannot be empty")
	}

	if len(documents) > MaxDocuments {
		return nil, fmt.Errorf(
			"maximum number of documents is %d",
			MaxDocuments,
		)
	}

	// Validate unique IDs and prepare documents for the reranker API.
	seen := make(map[int64]struct{}, len(documents))
	rerankDocuments := make([]Document, len(documents))

	for i := range documents {
		doc := documents[i]

		if _, exists := seen[doc.Id]; exists {
			return nil, fmt.Errorf(
				"duplicate document ID: %d",
				doc.Id,
			)
		}

		seen[doc.Id] = struct{}{}

		if len([]rune(doc.Title+"\n"+doc.URL+"\n"+doc.Body)) > MaxDocumentTextLength {
			rerankDocuments[i] = Document{
				ID: doc.Id,
				Text: truncateRunes(
					doc.Title+"\n"+doc.URL+"\n"+doc.Body,
					MaxDocumentTextLength,
				),
			}
			continue
		}

		rerankDocuments[i] = Document{
			ID:    doc.Id,
			Title: doc.Title,
			URL:   doc.URL,
			Text:  doc.Body,
		}
	}

	payload := RerankRequest{
		Query:     query,
		Documents: rerankDocuments,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal rerank request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/rerank",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create rerank request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call reranker API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Detail string `json:"detail"`
		}

		_ = json.NewDecoder(resp.Body).Decode(&apiErr)

		if apiErr.Detail == "" {
			apiErr.Detail = http.StatusText(resp.StatusCode)
		}

		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    apiErr.Detail,
		}
	}

	var result RerankResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode rerank response: %w", err)
	}

	return &result, nil
}

// Health checks whether the reranker model is ready.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/health",
		nil,
	)
	if err != nil {
		return fmt.Errorf("create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("check reranker health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return nil
}

func truncateRunes(text string, maxLength int) string {
	runes := []rune(text)

	if len(runes) <= maxLength {
		return text
	}

	return string(runes[:maxLength])
}
