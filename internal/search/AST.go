package search

import (
	"github.com/GabrielMoody/Cariin/internal/index"
)

type Query interface {
	Evaluate(index index.InvertedIdx) []int64
}

type TermQuery struct {
	Term string
}

type AndQuery struct {
	Left  Query
	Right Query
}

type OrQuery struct {
	Left  Query
	Right Query
}

type NotQuery struct {
	Term string
}

func (q TermQuery) Evaluate(index index.InvertedIdx) []int64 {
	return index[q.Term]
}

func (q AndQuery) Evaluate(index index.InvertedIdx) []int64 {
	left := q.Left.Evaluate(index)
	right := q.Right.Evaluate(index)

	return intersection(left, right)
}

func (q OrQuery) Evaluate(index index.InvertedIdx) []int64 {
	left := q.Left.Evaluate(index)
	right := q.Right.Evaluate(index)

	return union(left, right)
}

func (q NotQuery) Evaluate(index index.InvertedIdx) []int64 {
	var result []int64

	for range index {
		if v, ok := index[q.Term]; !ok {
			result = append(result, v...)
		}
	}

	return result
}

func intersection(a, b []int64) []int64 {
	set := make(map[int64]bool)

	for _, v := range a {
		set[v] = true
	}

	var result []int64

	for _, v := range b {
		if set[v] {
			result = append(result, v)
			set[v] = false
		}
	}

	return result
}

func union(a, b []int64) []int64 {
	set := make(map[int64]bool)

	for _, v := range a {
		set[v] = true
	}
	for _, v := range b {
		set[v] = true
	}

	var result []int64

	for v := range set {
		result = append(result, v)
	}
	return result
}
