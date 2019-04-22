package main

import "github.com/rivo/tview"

type Tabulation struct {
	Primitives   []*tview.Primitive
	CurrentIndex int
	OnChanged    func(primitive *tview.Primitive)
}

func (t *Tabulation) Focus(primitive *tview.Primitive) {
	index := -1
	for i := 0; i < len(t.Primitives); i++ {
		if t.Primitives[i] == primitive {
			index = i
		}
	}

	if index >= 0 {
		t.CurrentIndex = index
		if t.OnChanged != nil {
			t.OnChanged(primitive)
		}
	}
}

func (t *Tabulation) Next() {
	if len(t.Primitives) > 1 {
		if t.CurrentIndex >= len(t.Primitives)-1 {
			t.CurrentIndex = 0
		} else {
			t.CurrentIndex++
		}
		if t.OnChanged != nil {
			t.OnChanged(t.Primitives[t.CurrentIndex])
		}
	}
}

func (t *Tabulation) Prev() {
	if len(t.Primitives) > 1 {
		if t.CurrentIndex < 1 {
			t.CurrentIndex = len(t.Primitives) - 1
		} else {
			t.CurrentIndex--
		}
		if t.OnChanged != nil {
			t.OnChanged(t.Primitives[t.CurrentIndex])
		}
	}
}
