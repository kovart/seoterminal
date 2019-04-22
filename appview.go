package main

import (
	"fmt"
	"sort"
	"strings"
)

func (app *App) UpdateView(){
	app.UpdateClusterTree()
	app.UpdateStatusBar()
	app.UpdateKeywordList()

	app.Primitives.Input.SetText(app.State.Temp.Keyword)
}

func (app *App) UpdateClusterTree(){
	app.Primitives.ClusterTree.SetRoot(app.State.Temp.RootNode)
}

func (app *App) UpdateStatusBar() {
	statusBar := app.Primitives.StatusBar
	statusBar.Clear()
	//_, _, width, _ := statusBar.GetRect()

	clusterInfo := fmt.Sprintf("Всего запросов: %v | В корневом: %v | В текущем: %v", len(app.State.Project.Rows), len(app.State.Temp.RootNode.Rows), len(app.State.Temp.SelectedNode.Rows))

	fmt.Fprintf(statusBar, clusterInfo)
}

func (app *App) SetStatusBarText(msg string){
	statusBar := app.Primitives.StatusBar
	statusBar.Clear()
	fmt.Fprintf(statusBar, msg)
}

func (app *App) UpdateKeywordList(){

	list := app.Primitives.KeywordList
	list.Clear()
	sortRowsByStrongVolume(app.State.Temp.SelectedNode.Rows)
	var keywords []KeywordListItem
	for _, row := range app.State.Temp.SelectedNode.Rows {
		keyword := KeywordListItem{ Text:row.Keyword, Volume:row.Frequency, StrongVolume: row.StrongFrequency }
		keywords = append(keywords, keyword)
	}
	list.SetKeywords(keywords)
	//if app.State.Temp.SelectedNode != app.State.Temp.RootNode {
	//	list.SetTitle(strings.Join(app.State.Temp.SelectedNode.GetClusterNames(), " "))
	//} else {
	//	list.SetTitle(app.State.Temp.SelectedNode.GetFullName())
	//}
	list.SetTitle(app.State.Temp.SelectedNode.GetFullName())
}

func (app *App) ExpandAndSelect(clusterName string) {
	root := app.State.Temp.RootNode
	clusterWords := strings.Fields(clusterName)

	closest := root
	closestWords := strings.Fields(closest.GetFullName())
	closestSimilarity := orderSimilarity(clusterWords, closestWords)

	closest.Expand()
	var nodes = root.children
	if len(closest.children) == 0 {
		nodes = root.GenerateChildren(app.State.Temp.CachedClusters)
		sortClusterNodes(nodes)
		root.SetChildren(nodes)
	}
	for len(nodes) > 0 {
		node := nodes[0]
		nodes = nodes[1:]

		nodeWords := strings.Fields(node.GetFullName())
		nodeSimilarity := orderSimilarity(clusterWords, nodeWords)
		if nodeSimilarity == len(clusterWords) {
			closest = node
			break
		} else if len(nodeWords) >= len(clusterWords){
			continue
		} else if nodeSimilarity > closestSimilarity {
			closestSimilarity = nodeSimilarity
			closest = node
			nodes = closest.GenerateChildren(app.State.Temp.CachedClusters)
			sortClusterNodes(nodes)
			closest.SetChildren(nodes)
			closest.Expand()
		}
	}

	app.Primitives.ClusterTree.SetCurrentNode(closest)
}

func orderSimilarity(a, b[]string) int {
	similarity := 0
	for i:=0; i<len(a) && i <len(b);i++ {
		if a[i] != b[i] {
			break
		}
		similarity++
	}
	return similarity
}


func sortClusterNodes(nodes []*ClusterNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return len(nodes[i].Rows) > len(nodes[j].Rows)
	})
}

func sortRowsByVolume(rows []*Row) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Frequency > rows[j].Frequency
	})
}
func sortRowsByStrongVolume(rows []*Row) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].StrongFrequency > rows[j].StrongFrequency
	})
}