package index

import (
	"sync"

	"github.com/GabrielMoody/Cariin/internal/documents"
)

type Posting struct {
	DocId     int64
	Positions []int
}

type InvertedIdx map[string][]Posting

var (
	Idx  *InvertedIdx
	once sync.Once
)

func Get() *InvertedIdx {
	once.Do(func() {
		Idx = BuildInvertedIdx()
	})

	return Idx
}

func BuildInvertedIdx() *InvertedIdx {
	return Build(documents.Documents)
}
