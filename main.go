package main

import (
	"flag"
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"os"
)

// To run this app:
// $ tool -p ./ProjectName
// $ tool -p ./ProjectName -f ./keyword.csv

func main() {
	project := loadProjectCLI()
	// Nil means there are no things to do
	if project == nil {
		return
	}

	app := App{}
	app.State.Project = *project
	app.State.Temp.CachedClusters = make(map[string]*Cluster)
	app.UI = tview.NewApplication()

	root := initPrimitives(&app)

	app.UpdateView()
	if err := app.UI.SetRoot(root, true).SetFocus(root).Run(); err != nil {
		panic(err)
	}
}

func loadProjectCLI() (project *Project) {
	pFlag := flag.String("p", "", "Project name")
	fFlag := flag.String("f", "", "Keywords csv file")
	update := flag.Bool("update", false, "Re-cut all keywords in history.txt")
	helpFlag := flag.Bool("help", false, "Project name")

	flag.Parse()

	if *helpFlag {
		showHelp()
		return nil
	}

	if *pFlag == "" {
		panic("-p flag must be specified")
	}

	if *fFlag == "" {
		project = LoadProject(*pFlag, *update)
	} else {
		project = CreateProject(*pFlag, *fFlag, *update)
	}

	return
}

func initPrimitives(app *App) (root *tview.Pages) {
	currentPage := "Main"

	searchInput := tview.NewInputField()
	statusBar := tview.NewTextView()
	clusterTree := NewKeywordTreeView()
	keywordList := NewKeywordList([]KeywordListItem{})

	app.Primitives.Input = searchInput
	app.Primitives.ClusterTree = clusterTree
	app.Primitives.KeywordList = keywordList
	app.Primitives.StatusBar = statusBar

	app.SearchKeyword("")

	// ClusterTreeView
	// Initialization of primitive

	statusBar.SetDynamicColors(true)
	statusBar.SetBorderPadding(0, 0, 1, 1)

	keywordList.SetBorderPadding(0, 0, 1, 0).SetBorder(true).SetTitle("Запросы")

	clusterTree.SetBorder(true).SetTitle("Все слова")
	clusterTree.SetNavigatedFunc(func(node *ClusterNode) {
		app.State.Temp.SelectedNode = node
		app.UpdateKeywordList()
		app.UpdateStatusBar()
	})
	clusterTree.SetSelectedFunc(func(node *ClusterNode) {
		if len(node.children) == 0 {
			children := node.GenerateChildren(app.State.Temp.CachedClusters)
			sortClusterNodes(children)
			node.SetChildren(children)
		}
		node.IsExpanded = !node.IsExpanded
	})
	clusterTree.SetControlFunc(func(node *ClusterNode, key *tcell.EventKey) {
		if key.Key() == tcell.KeyCtrlK {
			app.SearchKeyword(node.Name)
			app.UpdateView()
			app.SetStatusBarText("Теперь корневой запрос: " + node.Name)
		} else if key.Key() == tcell.KeyCtrlA {
			app.SearchKeyword("")
			app.UpdateView()
			app.SetStatusBarText("Теперь показываются все слова")
		} else if key.Key() == tcell.KeyCtrlD {
			cut := filterRows(node.Name, app.State.Project.Rows)
			app.ProcessOperation(node.Name, cut, OperationRemove)
			if len(node.Neighbors) > 0 {
				app.ExpandAndSelect(node.Neighbors[0].GetFullName())
			}
			app.UpdateView()
			app.SetStatusBarText(fmt.Sprintf("Удален корневой кластер:[red] %s", node.Name))
		} else if key.Rune() == '/' {
			cut := filterRows(node.Name, app.State.Project.Rows)
			app.ProcessOperation(node.Name, cut, OperationSilentRemove)
			if len(node.Neighbors) > 0 {
				app.ExpandAndSelect(node.Neighbors[0].GetFullName())
			}
			app.UpdateView()
			app.SetStatusBarText(fmt.Sprintf("Удален без извлечения кластер:[red] %s", node.Name))
		} else if key.Key() == tcell.KeyCtrlS {
			cut := filterRows(node.Name, app.State.Project.Rows)
			app.ProcessOperation(node.Name, cut, OperationAdd)
			if len(node.Neighbors) > 0 {
				app.ExpandAndSelect(node.Neighbors[0].GetFullName())
			}
			app.UpdateView()
			app.SetStatusBarText(fmt.Sprintf("Сохранен корневой кластер:[green] %s", node.Name))
		} else if key.Rune() == '+' || key.Key() == tcell.KeyCtrlSpace {
			app.ProcessOperation(node.GetFullName(), node.Rows, OperationAdd)
			if len(node.Neighbors) > 0 {
				app.ExpandAndSelect(node.Neighbors[0].GetFullName())
			}
			app.UpdateView()
			app.SetStatusBarText(fmt.Sprintf("Сохранен кластер:[green] %s", node.GetFullName()))
		} else if key.Key() == tcell.KeyBackspace2 || key.Rune() == '-' {
			app.ProcessOperation(node.GetFullName(), node.Rows, OperationRemove)
			if len(node.Neighbors) > 0 {
				app.ExpandAndSelect(node.Neighbors[0].GetFullName())
			}
			app.UpdateView()
			app.SetStatusBarText(fmt.Sprintf("Удален кластер:[red] %s", node.GetFullName()))
		}
	})

	searchInput.SetBorder(true)
	searchInput.SetFieldBackgroundColor(tcell.ColorDefault)
	searchInput.SetLabel(" Запрос: ")
	searchInput.SetLabelColor(tcell.ColorWhite)
	searchInput.SetFieldBackgroundColor(0x586E75)
	searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.SearchKeyword(app.Primitives.Input.GetText())
			app.UpdateView()
		}
	})

	mainFlex := tview.NewFlex()
	mainFlex.SetDirection(tview.FlexColumn).
		AddItem(clusterTree, 0, 7, true).
		AddItem(keywordList, 0, 14, true)

	grid := tview.NewGrid().SetRows(3, 0, 1).SetColumns(30, 0, 12, 12)
	grid.SetBackgroundColor(0x2E3436)
	grid.AddItem(searchInput, 0, 0, 1, 4, 0, 0, true)
	grid.AddItem(mainFlex, 1, 0, 1, 4, 0, 0, false)
	grid.AddItem(statusBar, 2, 0, 1, 4, 0, 0, false)

	pages := tview.NewPages()
	pages.AddPage("Main", grid, true, true)

	tabPrimitives := []tview.Primitive{
		searchInput,
		clusterTree,
		keywordList,
	}

	currentPrimitive := 0
	app.UI.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if currentPage == "Main" {
			if event.Key() == tcell.KeyTab {
				if currentPrimitive >= len(tabPrimitives)-1 {
					currentPrimitive = 0
				} else {
					currentPrimitive++
				}
				app.UI.SetFocus(tabPrimitives[currentPrimitive])
			} else if event.Key() == tcell.KeyBacktab {
				if currentPrimitive < 1 {
					currentPrimitive = len(tabPrimitives) - 1
				} else {
					currentPrimitive--
				}
				app.UI.SetFocus(tabPrimitives[currentPrimitive])
			} else if event.Modifiers()&tcell.ModAlt != 0 && (event.Rune() == 'h' || event.Rune() == 'р') {
				history := historyList(app, func(operation *KeywordOperation) {
					pages.SwitchToPage("Main")
					currentPage = "Main"
					pages.RemovePage("History")
					if operation != nil {
						app.State.Project.History.SetOperation(operation)
						go app.State.Project.SaveHistory()
						app.State.Project.ApplyHistory()
						app.SearchKeyword(app.State.Temp.Keyword)
						app.UI.SetFocus(tabPrimitives[currentPrimitive])
						app.UpdateView()
						app.UI.Draw()
					}
				})
				pages.AddAndSwitchToPage("History", history, true)
				currentPage = "History"
				return tcell.NewEventKey(tcell.KeyNUL, 0, tcell.ModNone)
			}
		}

		return event
	})

	return pages
}

func historyList(app *App, done func(operation *KeywordOperation)) *SimpleList {
	list := NewSimpleList()
	list.SetBorder(true).SetTitle("История").SetBorderPadding(0, 0, 1, 1)
	for i := len(app.State.Project.History.Operations) - 1; i >= 0; i-- {
		index := i
		oper := app.State.Project.History.Operations[i]
		var color string = "[green]"
		if oper.Operation != OperationAdd {
			color = "[red]"
		}
		var pointer string
		if i == app.State.Project.History.CurrentStateIndex {
			pointer = " <-- текущий"
		}
		var prefix string
		if oper.Operation == OperationSilentRemove {
			prefix = "-- "
		}
		list.AddItem(fmt.Sprintf("%s%v. %s%s [white]%s", color, i+1, prefix, oper.Keyword, pointer), func() {
			done(app.State.Project.History.Operations[index])
		})
	}
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			done(nil)
		}
		return event
	})
	return list
}

func showHelp() {
	fmt.Println()
	fmt.Println("Примеры запуска:")
	fmt.Println(" -p projects/spina")
	fmt.Println("  Запуск существующего проекта")
	fmt.Println(" -p projects/spina -f csv/all.csv")
	fmt.Println("  Создание проекта с указанием источника запросов")
	fmt.Println(" -p projects/spina -update")
	fmt.Println("  Открытие проекта с пересохранением ключевых слов в файлы")
	fmt.Println("\nГде:")
	fmt.Println(" \"-p\" — путь к проекту")
	fmt.Println(" \"-f\" — путь к файлу с запросами, по которому создастся проект")
	fmt.Println(" \"-update\" — комманда вырезать все ключевые слова из файла history.txt")
	fmt.Println()
	fmt.Println("Управление деревом кластеров:")
	fmt.Println(" +             — добавить кластер")
	fmt.Println(" Backspace/-   — удалить кластер")
	fmt.Println(" Ctrl+D        — удалить рутовый кластер")
	fmt.Println(" Ctrl+S        — сохранить рутовый кластер")
	fmt.Println(" /             — удалить рутовый кластер без вырезания")
	fmt.Println()
}

// Currently unused code
// I've used this code to graphically ask the target csv file
// It required many lines of code so I decided to cut this feature

func askAboutFile(files []os.FileInfo, done func(file os.FileInfo)) tview.Primitive {
	var alphabet = []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle("Выберите файл с запросами")
	for i, f := range files {
		list.AddItem(f.Name(), fmt.Sprintf("Размер: %v", f.Size()), alphabet[i], func() {
			done(f)
		})
	}
	return list
}
