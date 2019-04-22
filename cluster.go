package main

import (
	"sort"
	"strings"
)

type Cluster struct {
	Rows []*Row
	Hash string
}

func NewCluster(name string, rows []*Row, parent *Cluster) *Cluster {
	var hash string
	if parent != nil {
		words := strings.Fields(parent.Hash)
		words = append(words, name)
		sort.Strings(words)
		hash = strings.Join(words, " ")
	} else {
		hash = name
	}
	return &Cluster{Hash: hash, Rows: rows}
}

func (c *Cluster) GenerateSubClusters(existedClusterNodes map[string]*Cluster, minKeywords uint) map[string]*Cluster {
	// Ugly fix
	if existedClusterNodes == nil {
		existedClusterNodes = make(map[string]*Cluster)
	}

	parent := c
	rows := c.Rows
	wordMap := make(map[string][]*Row)
	for _, row := range rows {
		words := strings.Fields(row.NormalizedKeyword)
		for _, word := range words {
			val, ok := wordMap[word]
			if ok {
				wordMap[word] = append(val, row)
			} else {
				wordMap[word] = []*Row{row}
			}
		}
	}

	// RemoveNode parent words
	excludeWords := strings.Fields(parent.Hash)
	for _, v := range excludeWords {
		delete(wordMap, v)
	}

	getHash := func(words []string) string {
		sort.Strings(words)
		return strings.Join(words, " ")
	}

	// Generate clusters slice
	clusters := make(map[string]*Cluster, len(wordMap))
	for k := range wordMap {
		hash := getHash(append(excludeWords, k))
		var cluster *Cluster
		if v, ok := existedClusterNodes[hash]; ok == true {
			cluster = v
		} else {
			cluster = &Cluster{Rows: wordMap[k], Hash: hash}
			existedClusterNodes[hash] = cluster
		}
		if len(cluster.Rows) >= int(minKeywords) {
			clusters[k] = cluster
		}
	}
	return clusters
}


