package documents

import (
	"time"

	"github.com/GabrielMoody/Cariin/internal/db"
)

type Document struct {
	Id    int64
	URL   string
	Title string
	Body  string
}

type T_Document struct {
	Id          int64
	URL         string
	Domain      string
	Title       string
	Description string
	Body        string

	CrawledAt  time.Time
	StatusCode int
}

var Documents = map[int64]Document{}

func Init() {
	db := db.Init()
	Documents, _ = GetDocuments(db)
}
