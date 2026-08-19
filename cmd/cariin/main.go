package main

import (
	"fmt"

	"github.com/GabrielMoody/Cariin/internal/documents"
	"github.com/GabrielMoody/Cariin/internal/index"
)

func main() {
	idx := index.Build(documents.Documents)

	fmt.Println(idx)
}
