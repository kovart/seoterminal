package main

import (
	"strings"
)

type Row struct {
	NormalizedKeyword   string `csv:"Лемма"`
	Keyword             string `csv:"Ключевое слово"`
	Frequency           uint32 `csv:"Широкая частотность"`
	StrongFrequency     uint32 `csv:"Строгая частотность"`
}


func filterRows(keyword string, rows []*Row) []*Row {
	wordMap := generateWordMap(rows)
	keywordWords := strings.Fields(keyword)

	if len(keywordWords) == 0 {
		return rows
	}

	var goodRows []*Row
	firstWord := keywordWords[0]
	if keywordRows, ok := wordMap[firstWord]; ok {
		for i, row := range keywordRows {
			rowWords := strings.Fields(row.NormalizedKeyword)
			if contains(rowWords, keywordWords[1:]) {
				goodRows = append(goodRows, keywordRows[i])
			}
		}
	}

	return goodRows
}

func ApplyHistory(rows []*Row, history *History) (ret []*Row) {
	ret = rows
	var keywords []string
	for i, operation := range history.Operations {
		if i > history.CurrentStateIndex {
			break
		}
		keywords = append(keywords, operation.Keyword)
	}
	ret = removeRows(keywords, rows)
	return
}


func removeRows(keywords []string, rows []*Row) []*Row {
	// [
	// 	"грыжа": [Row1, Row2],
	//	"боль" : [Row2, Row3]
	// ]
	wordMap := generateWordMap(rows)


	rowMap := convertRowsToMap(rows)
	for _, keyword := range keywords {
		keywordWords := strings.Fields(keyword)

		if len(keywordWords) == 0 {
			continue
		}

		firstWord := keywordWords[0]
		if keywordRows, ok := wordMap[firstWord]; ok {
			var leftRows []*Row
			for i, row := range keywordRows {
				rowWords := strings.Fields(row.NormalizedKeyword)
				if contains(rowWords, keywordWords[1:]) {
					delete(rowMap, keywordRows[i])
				} else {
					leftRows = append(leftRows, keywordRows[i])
				}
			}
			wordMap[firstWord] = leftRows
		}
	}

	rows = convertMapToRows(rowMap)
	return rows
}

func contains(arr []string, subArray []string) bool {
	if len(subArray) == 0 {
		return true
	}
	if len(subArray) > len(arr) {
		return false
	}

	arrMap := stringsToMap(arr)
	for _, sub := range subArray {
		if _, ok := arrMap[sub]; !ok {
			return false
		}
	}
	return true
}

func stringsToMap(arr []string) map[string]struct{} {
	m := make(map[string]struct{})

	for _, s := range arr {
		m[s] = struct{}{}
	}

	return m
}

func generateWordMap(rows []*Row) map[string][]*Row{
	wordMap := make(map[string][]*Row)

	// Generating map of rows for every word
	for i, r := range rows {
		words := strings.Fields(r.NormalizedKeyword)
		for _, word := range words {
			wordMap[word] = append(wordMap[word], rows[i])
		}
	}
	return wordMap
}