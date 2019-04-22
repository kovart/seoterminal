package main

import (
	"fmt"
	"github.com/jszwec/csvutil"
	"gopkg.in/cheggaaa/pb.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	ProjectRemainFile   = "remains.csv"
	ProjectOriginalFile = "original.csv"
	ProjectHistoryFile  = "history.txt"
	ProjectClustersDir  = "clusters"
	ProjectRemovedDir   = "removed"
)

type Project struct {
	Rows        []*Row
	InitialRows []*Row

	History *History
	Paths   ProjectPaths
}

type ProjectPaths struct {
	Dir          string // project location
	OriginalFile string
	RemainsFile  string
	HistoryFile  string
	ClustersDir  string
	RemovedDir   string
}

func CreateProject(path, csvFile string, createKeywordFiles bool) *Project {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return LoadProject(path, createKeywordFiles)
	}

	paths := resolveProjectPaths(path)
	err := os.MkdirAll(paths.Dir, 0777)
	check(err)
	err = copyFile(csvFile, paths.OriginalFile)
	check(err)
	err = copyFile(csvFile, paths.RemainsFile)
	check(err)

	rows := LoadRows(paths.OriginalFile)

	return &Project{
		Rows:        rows,
		InitialRows: rows,
		Paths:       *paths,
		History:     &History{},
	}
}

// Create and return reference to project structure
// Parameter bar can be nil
func LoadProject(path string, createHistoryFiles bool) *Project {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		panic("Project does not exist")
	}

	project := Project{}
	project.Paths = *resolveProjectPaths(path)
	project.InitialRows = LoadRows(project.Paths.OriginalFile)
	project.History = LoadHistory(project.Paths.HistoryFile)
	//if createHistoryFiles {
	//	project.History.CurrentStateIndex = len(project.History.Operations) - 1
	//}

	if createHistoryFiles {
		project.Rows = project.SaveAndApplyOperationKeywords()
	} else {
		project.Rows = ApplyHistory(project.InitialRows, project.History)
	}
	go project.Save()

	return &project
}

func (p *Project) Save() {
	mutex := &sync.Mutex{}
	data, err := csvutil.Marshal(p.Rows[:])
	check(err)
	mutex.Lock()
	err = ioutil.WriteFile(p.Paths.RemainsFile, data, 0777)
	check(err)

	p.History.Save(p.Paths.HistoryFile)
	mutex.Unlock()
	check(err)
}

func (p *Project) SaveHistory() {
	mutex := &sync.Mutex{}
	mutex.Lock()
	p.History.Save(p.Paths.HistoryFile)
	mutex.Unlock()
}

func (p *Project) RemoveRows(rows []*Row) {
	rowMap := convertRowsToMap(p.Rows)

	for i := range rows {
		delete(rowMap, rows[i])
	}

	p.Rows = convertMapToRows(rowMap)
}



func (p *Project) ApplyHistory() {
	rows := p.InitialRows
	p.Rows = ApplyHistory(rows, p.History)
}

func (p *Project) SaveAndApplyOperationKeywords() []*Row {
	wordMap := generateWordMap(p.InitialRows)
	rowMap := convertRowsToMap(p.InitialRows)

	fmt.Println("Cutting operations...")
	bar := pb.StartNew(len(p.History.Operations))
	for opIndex, op := range p.History.Operations {
		// Rows with the current operation keyword
		var operatedRows []*Row
		keywordWords := strings.Fields(op.Keyword)
		if len(keywordWords) == 0 {
			continue
		}

		// Row search optimization
		firstWord := keywordWords[0]
		if keywordRows, ok := wordMap[firstWord]; ok {
			var leftRows []*Row
			for i, row := range keywordRows {
				rowWords := strings.Fields(row.NormalizedKeyword)
				if contains(rowWords, keywordWords[1:]) {
					operatedRows = append(operatedRows, keywordRows[i])
					if opIndex <= p.History.CurrentStateIndex  {
						delete(rowMap, keywordRows[i])
					}
				} else {
					leftRows = append(leftRows, keywordRows[i])
				}
			}
			wordMap[firstWord] = leftRows
		}

		switch op.Operation {
		case OperationAdd:
			SaveRows(operatedRows, filepath.Join(p.Paths.ClustersDir, op.Keyword+".csv"))
		case OperationRemove:
			SaveRows(operatedRows, filepath.Join(p.Paths.RemovedDir, op.Keyword+".csv"))
		}
		bar.Increment()
	}
	bar.Finish()

	return convertMapToRows(rowMap)
}

func resolveProjectPaths(path string) *ProjectPaths {
	absPath, err := filepath.Abs(path)
	check(err)
	project := ProjectPaths{}
	project.Dir = absPath
	project.OriginalFile = filepath.Join(project.Dir, ProjectOriginalFile)
	project.ClustersDir = filepath.Join(project.Dir, ProjectClustersDir)
	project.RemainsFile = filepath.Join(project.Dir, ProjectRemainFile)
	project.HistoryFile = filepath.Join(project.Dir, ProjectHistoryFile)
	project.RemovedDir = filepath.Join(project.ClustersDir, ProjectRemovedDir)
	return &project
}

func SaveRows(rows []*Row, path string) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0777)
	data, err := csvutil.Marshal(rows)
	check(err)
	err = ioutil.WriteFile(path, data, 0777)
	mutex.Unlock()
	check(err)
}

func LoadRows(file string) []*Row {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Can't read %s\n", file)
		panic(err)
	}
	var rows []Row
	if err := csvutil.Unmarshal(data, &rows); err != nil {
		panic("Can't parse csv file")
	}

	pointers := make([]*Row, len(rows))
	for i := 0; i < len(rows); i++ {
		pointers[i] = &rows[i]
	}

	return pointers
}

// Utils

func convertRowsToMap(rows []*Row) map[*Row]struct{} {
	m := make(map[*Row]struct{}, len(rows))
	for i := range rows {
		m[rows[i]] = struct{}{}
	}
	return m
}

func convertMapToRows(m map[*Row]struct{}) []*Row {
	rows := make([]*Row, len(m))
	i := 0
	for val := range m {
		rows[i] = val
		i++
	}
	return rows
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
