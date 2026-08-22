package index

import (
	"sync"

	"github.com/GabrielMoody/Cariin/internal/documents"
)

type InvertedIdx map[string][]Posting

type Posting struct {
	DocId     int64
	Positions []int
}

type Index struct {
	All   *InvertedIdx
	Title *InvertedIdx
	Body  *InvertedIdx
	Url   *InvertedIdx
}

var (
	Idx  *Index
	once sync.Once
)

func Get() *Index {
	once.Do(func() {
		Idx = BuildInvertedIdx()
	})

	return Idx
}

func BuildInvertedIdx() *Index {
	indexes := &Index{}

	indexes.Title = Build(documents.Documents, func(doc documents.Document) string {
		return doc.Title
	})

	indexes.Body = Build(documents.Documents, func(doc documents.Document) string {
		return doc.Body
	})

	indexes.Url = Build(documents.Documents, func(doc documents.Document) string {
		return doc.URL
	})

	indexes.All = Build(documents.Documents, func(doc documents.Document) string {
		return doc.Title + " " + doc.Body
	})

	return indexes
}
