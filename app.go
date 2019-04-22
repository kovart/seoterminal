package main

import (
	"github.com/rivo/tview"
	"path/filepath"
)

type App struct {
	UI         *tview.Application
	State      AppState
	Primitives AppPrimitives
}

type AppState struct {
	Project    Project
	Tabulation Tabulation
	Temp       AppTempState
}

type AppTempState struct {
	Keyword string

	CachedClusters map[string]*Cluster

	RootNode     *ClusterNode
	SelectedNode *ClusterNode
}

type AppPrimitives struct {
	// Inputs
	Input *tview.InputField

	// Boxes
	ClusterTree *ClusterTreeView
	KeywordList *KeywordList

	// Indicators
	StatusBar *tview.TextView
}

func (app *App) SearchKeyword(keyword string) {
	var root *Cluster
	var node *ClusterNode
	if keyword == "" {
		root = NewCluster(keyword, app.State.Project.Rows, nil)
		node = NewClusterNode("Слова", root, true, nil)
	} else {
		cut := filterRows(keyword, app.State.Project.Rows)
		root = NewCluster(keyword, cut, nil)
		node = NewClusterNode(keyword, root, true, nil)
	}
	children := node.GenerateChildren(app.State.Temp.CachedClusters)
	sortClusterNodes(children)
	node.SetChildren(children)

	app.State.Temp.CachedClusters = nil
	app.State.Temp.SelectedNode = node
	app.State.Temp.RootNode = node
	app.State.Temp.Keyword = keyword
	app.Primitives.ClusterTree.SetRoot(node)
	app.Primitives.ClusterTree.SetCurrentNode(node)
}

func (app *App) ProcessOperation(keyword string, rows []*Row, operation int) {
	var path string
	if operation != OperationSilentRemove {
		if operation == OperationRemove {
			path = filepath.Join(app.State.Project.Paths.RemovedDir, keyword+".csv")
		} else if operation == OperationAdd {
			path = filepath.Join(app.State.Project.Paths.ClustersDir, keyword+".csv")
		} else {
			panic("passed operation is unknown")
		}

		go SaveRows(rows, path)
	}

	app.State.Project.RemoveRows(rows)
	app.State.Project.History.AddOperation(keyword, operation)
	app.SearchKeyword(app.State.Temp.Keyword)
	go app.State.Project.Save()
}
