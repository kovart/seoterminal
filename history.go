package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var lineRegexp = regexp.MustCompile(`((?P<operation>([-+/!]{1,2}))\s*(?P<text>.*))`)

const (
	OperationRemove = iota
	OperationSilentRemove
	OperationAdd
)

type History struct {
	Operations        []*KeywordOperation
	CurrentStateIndex int
}

type KeywordOperation struct {
	Keyword   string
	Operation int
}

func LoadHistory(path string) *History {
	history := History{}
	file, err := os.OpenFile(path, os.O_RDONLY, 0777)
	defer file.Close()
	if os.IsNotExist(err) {
		return &history
	}
	check(err)

	scanner := bufio.NewScanner(file)

	// Load current operation
	scanner.Scan()
	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine == "" {
		return &history
	}
	currentOperation := strings.TrimSpace(firstLine[1:])

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		match := lineRegexp.FindStringSubmatch(text)
		paramsMap := make(map[string]string)
		for i, name := range lineRegexp.SubexpNames() {
			if i > 0 && i <= len(match) {
				paramsMap[name] = strings.TrimSpace(match[i])
			}
		}

		operation, ok := paramsMap["operation"]
		if !ok {
			panic("invalid operation: " + text)
		}

		operationType := -1
		switch operation {
		case "-":
			operationType = OperationRemove
		case "--":
			operationType = OperationSilentRemove
		case "+":
			operationType = OperationAdd
		}

		text, ok = paramsMap["text"]
		if !ok {
			panic("invalid operation: " + text)
		}

		history.AddOperation(text, operationType)
	}
	history.Operations = removeDuplicates(history.Operations)
	if currentOperation == "" {
		history.CurrentStateIndex = len(history.Operations) - 1
	} else {
		for i, op := range history.Operations {
			if currentOperation == op.Keyword {
				history.CurrentStateIndex = i
				break
			}
		}
	}

	return &history
}

func (h *History) Save(path string) {
	if len(h.Operations) == 0 {
		return
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	defer file.Close()
	check(err)
	writer := bufio.NewWriter(file)
	fmt.Fprintf(writer, "= %s\n\n", h.Operations[h.CurrentStateIndex].Keyword)
	for _, operation := range h.Operations {
		var prefix string
		if operation.Operation == OperationRemove {
			prefix = "-"
		} else if operation.Operation == OperationSilentRemove {
			prefix = "--"
		} else if operation.Operation == OperationAdd {
			prefix = "+"
		}

		fmt.Fprintf(writer, "%s %s\n", prefix, operation.Keyword)
	}
	writer.Flush()
}

func (h *History) AddOperation(keyword string, operation int) {
	op := KeywordOperation{
		Keyword:   keyword,
		Operation: operation,
	}
	if len(h.Operations) == 0 || h.CurrentStateIndex == len(h.Operations)-1 {
		h.Operations = append(h.Operations, &op)
	} else {
		operations := h.Operations[:h.CurrentStateIndex+1]
		h.Operations = append(operations, &op)
	}
	h.CurrentStateIndex = len(h.Operations) - 1
}

func (h *History) SetOperation(operation *KeywordOperation) {
	for i := range h.Operations {
		if h.Operations[i] == operation {
			h.CurrentStateIndex = i
			break
		}
	}
}

func removeDuplicates(operations []*KeywordOperation) []*KeywordOperation {
	m := make(map[string]struct{}, len(operations))
	var clean []*KeywordOperation
	for i, op := range operations {
		if _, ok := m[op.Keyword]; !ok {
			m[op.Keyword] = struct{}{}
			clean = append(clean, operations[i])
		}
	}
	return clean
}
