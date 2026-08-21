package ast

import (
	"sort"

	"github.com/GabrielMoody/Cariin/internal/index"
)

type Query interface {
	Evaluate(idx index.InvertedIdx) []index.Posting
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
	Term  string
	Query Query
}

func (q TermQuery) Evaluate(idx index.InvertedIdx) []index.Posting {
	return idx[q.Term]
}

func (q AndQuery) Evaluate(idx index.InvertedIdx) []index.Posting {
	left := q.Left.Evaluate(idx)
	right := q.Right.Evaluate(idx)

	return intersection(left, right)
}

func (q OrQuery) Evaluate(idx index.InvertedIdx) []index.Posting {
	left := q.Left.Evaluate(idx)
	right := q.Right.Evaluate(idx)

	return union(left, right)
}

func (q NotQuery) Evaluate(idx index.InvertedIdx) []index.Posting {
	operand := q.Term
	var excluded []index.Posting
	if q.Query != nil {
		excluded = q.Query.Evaluate(idx)
	} else {
		excluded = idx[operand]
	}

	excludedSet := make(map[int64]bool, len(excluded))
	for _, posting := range excluded {
		excludedSet[posting.DocId] = true
	}

	allDocuments := make(map[int64]bool)
	for _, postings := range idx {
		for _, posting := range postings {
			allDocuments[posting.DocId] = true
		}
	}

	var result []index.Posting
	for id := range allDocuments {
		if !excludedSet[id] {
			result = append(result, index.Posting{DocId: id})
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].DocId < result[j].DocId })
	return result
}

func intersection(a, b []index.Posting) []index.Posting {
	set := make(map[int64]bool)

	for _, v := range a {
		set[v.DocId] = true
	}

	var result []index.Posting

	for _, v := range b {
		if set[v.DocId] {
			result = append(result, v)
			set[v.DocId] = false
		}
	}

	return result
}

func union(a, b []index.Posting) []index.Posting {
	set := make(map[int64]index.Posting)

	for _, v := range a {
		set[v.DocId] = v
	}
	for _, v := range b {
		if _, exists := set[v.DocId]; !exists {
			set[v.DocId] = v
		}
	}

	var result []index.Posting

	for _, v := range set {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DocId < result[j].DocId })
	return result
}
