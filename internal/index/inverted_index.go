package index

import (
	"sync"

	"github.com/GabrielMoody/Cariin/internal/documents"
)

type InvertedIdx map[string][]int64

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
