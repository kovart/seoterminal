package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	MovementNone = iota
	MovementUp
	MovementDown
)

type KeywordListItem struct {
	Text         string
	Volume       uint32
	StrongVolume uint32
}

type KeywordList struct {
	*tview.Box
	rows     []KeywordListItem
	yOffset  int
	movement int
}

func NewKeywordList(rows []KeywordListItem) *KeywordList {
	return &KeywordList{
		Box:      tview.NewBox(),
		rows:     rows,
		movement: MovementNone,
	}
}

func (r *KeywordList) SetKeywords(rows []KeywordListItem) {
	r.rows = rows
	r.yOffset = 0
}

func (r *KeywordList) Clear() {
	r.rows = nil
}

func (r *KeywordList) Draw(screen tcell.Screen) {
	r.Box.Draw(screen)
	x, y, width, height := r.GetInnerRect()

	if r.movement == MovementUp && r.yOffset > 0 {
		r.yOffset--
	}
	if r.movement == MovementDown && r.yOffset < (len(r.rows)-height) {
		r.yOffset++
	}
	r.movement = MovementNone

	volumeLength := 9
	strongVolumeLength := 6
	for index, row := range r.rows {
		if index < r.yOffset {
			continue
		}

		if index-r.yOffset >= height {
			break
		}

		text := row.Text
		volumeText := fmt.Sprintf(" │%"+strconv.Itoa(volumeLength)+"v", row.Volume)
		strongVolumeText := fmt.Sprintf(" │%"+strconv.Itoa(strongVolumeLength)+"v", row.StrongVolume)

		if width <= 0 || width-8 <= 0 {
			return
		}
		if width < 35 {
			volumeText = ""
		}
		if width < 25 {
			strongVolumeText = ""
		}

		length := utf8.RuneCountInString(row.Text) +
			utf8.RuneCountInString(volumeText) +
			utf8.RuneCountInString(strongVolumeText)
		if length > width {
			textLength := width - utf8.RuneCountInString(strongVolumeText) -
				utf8.RuneCountInString(volumeText) - 1
			text = string([]rune(text)[:textLength]) + "\u2026"
		} else {
			text = text + strings.Repeat(" ", width-length)
		}
		line := fmt.Sprintf(`%s[yellow]%s[yellow]%s[yellow]%s`, text, strongVolumeText, volumeText)
		tview.Print(screen, line, x, y+index-r.yOffset, width, tview.AlignLeft, tcell.ColorWhite)
	}
}

func Format(n uint32) string {
	in := strconv.FormatInt(int64(n), 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

// InputHandler returns the handler for this primitive.
func (t *KeywordList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Because the tree is flattened into a list only at drawing time, we also
		// postpone the (selection) movement to drawing time.
		switch key := event.Key(); key {
		case tcell.KeyDown:
			t.movement = MovementDown
		case tcell.KeyUp:
			t.movement = MovementUp
		}
	})
}
