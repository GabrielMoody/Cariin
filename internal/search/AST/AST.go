package ast

import (
	"sort"
	"strings"

	"github.com/GabrielMoody/Cariin/internal/index"
)

type Query interface {
	Evaluate(idx index.Index) []index.Posting
}

type TermQuery struct {
	Field string
	Term  string
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

func (q TermQuery) Evaluate(idx index.Index) []index.Posting {
	fieldIndex := selectFieldIndex(idx, q.Field)
	terms := strings.Fields(strings.ToLower(q.Term))
	if len(terms) <= 1 {
		return fieldIndex[terms[0]]
	}

	postingsByDocument := make(map[int64]index.Posting)
	for _, posting := range fieldIndex[terms[0]] {
		postingsByDocument[posting.DocId] = posting
	}

	for _, term := range terms[1:] {
		termPostings := make(map[int64]index.Posting)
		for _, posting := range fieldIndex[term] {
			termPostings[posting.DocId] = posting
		}

		for docID := range postingsByDocument {
			posting, exists := termPostings[docID]
			if !exists {
				delete(postingsByDocument, docID)
				continue
			}

			previous := postingsByDocument[docID]
			if hasAdjacentPosition(previous.Positions, posting.Positions) {
				continue
			}
		}
	}

	var result []index.Posting
	for _, posting := range postingsByDocument {
		result = append(result, posting)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DocId < result[j].DocId })
	return result
}

func selectFieldIndex(idx index.Index, field string) index.InvertedIdx {
	var selected *index.InvertedIdx
	switch field {
	case "title":
		selected = idx.Title
	case "body":
		selected = idx.Body
	case "url":
		selected = idx.Url
	default:
		selected = idx.All
	}

	if selected == nil {
		return index.InvertedIdx{}
	}
	return *selected
}

func hasAdjacentPosition(left, right []int) bool {
	for _, leftPosition := range left {
		for _, rightPosition := range right {
			if rightPosition == leftPosition+1 || leftPosition == rightPosition+1 {
				return true
			}
		}
	}
	return false
}

func (q AndQuery) Evaluate(idx index.Index) []index.Posting {
	left := q.Left.Evaluate(idx)
	right := q.Right.Evaluate(idx)

	return intersection(left, right)
}

func (q OrQuery) Evaluate(idx index.Index) []index.Posting {
	left := q.Left.Evaluate(idx)
	right := q.Right.Evaluate(idx)

	return union(left, right)
}

func (q NotQuery) Evaluate(idx index.Index) []index.Posting {
	operand := q.Term
	var excluded []index.Posting
	if q.Query != nil {
		excluded = q.Query.Evaluate(idx)
	} else {
		excluded = (*idx.All)[operand]
	}

	excludedSet := make(map[int64]bool, len(excluded))
	for _, posting := range excluded {
		excludedSet[posting.DocId] = true
	}

	allDocuments := make(map[int64]bool)
	for _, postings := range *idx.All {
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

// func evaluateFieldTerm(q TermQuery, idx index.Index) []index.Posting {
// 	var result []index.Posting

// 	switch q.Field {
// 	case "":
// 		result = append(result, (*idx.All)[q.Term]...)
// 	}
// }

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
