// This is a changed clone of Tview Component (primitive)

package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// listItem represents one item in a SimpleList.
type listItem struct {
	MainText string // The main text of the list item.
	Selected func() // The optional function which is called when the item is selected.
}

// SimpleList displays rows of items, each of which can be selected.
//
// See https://github.com/rivo/tview/wiki/List for an example.
type SimpleList struct {
	*tview.Box

	// The items of the list.
	items []*listItem

	// The index of the currently selected item.
	currentItem int

	// The item main text color.
	mainTextColor tcell.Color

	// The text color for selected items.
	selectedTextColor tcell.Color

	// The background color for selected items.
	selectedBackgroundColor tcell.Color

	// An optional function which is called when the user has navigated to a list
	// item.
	changed func(index int)

	// An optional function which is called when a list item was selected. This
	// function will be called even if the list item defines its own callback.
	selected func(index int)

	// An optional function which is called when the user presses the Escape key.
	done func()
}

// NewSimpleList returns a new form.
func NewSimpleList() *SimpleList {
	return &SimpleList{
		Box:                     tview.NewBox(),
		mainTextColor:           tview.Styles.PrimaryTextColor,
		selectedTextColor:       tview.Styles.PrimitiveBackgroundColor,
		selectedBackgroundColor: tview.Styles.PrimaryTextColor,
	}
}

// SetCurrentItem sets the currently selected item by its index. This triggers
// a "changed" event.
func (l *SimpleList) SetCurrentItem(index int) *SimpleList {
	l.currentItem = index
	if l.currentItem < len(l.items) && l.changed != nil {
		l.changed(l.currentItem)
	}
	return l
}

// GetCurrentItem returns the index of the currently selected list item.
func (l *SimpleList) GetCurrentItem() int {
	return l.currentItem
}

// SetMainTextColor sets the color of the items' main text.
func (l *SimpleList) SetMainTextColor(color tcell.Color) *SimpleList {
	l.mainTextColor = color
	return l
}

// SetSelectedTextColor sets the text color of selected items.
func (l *SimpleList) SetSelectedTextColor(color tcell.Color) *SimpleList {
	l.selectedTextColor = color
	return l
}

// SetSelectedBackgroundColor sets the background color of selected items.
func (l *SimpleList) SetSelectedBackgroundColor(color tcell.Color) *SimpleList {
	l.selectedBackgroundColor = color
	return l
}

// SetChangedFunc sets the function which is called when the user navigates to
// a list item. The function receives the item's index in the list of items
// (starting with 0), its main text, secondary text, and its shortcut rune.
//
// This function is also called when the first item is added or when
// SetCurrentItem() is called.
func (l *SimpleList) SetChangedFunc(handler func(int)) *SimpleList {
	l.changed = handler
	return l
}

// SetSelectedFunc sets the function which is called when the user selects a
// list item by pressing Enter on the current selection. The function receives
// the item's index in the list of items (starting with 0), its main text,
// secondary text, and its shortcut rune.
func (l *SimpleList) SetSelectedFunc(handler func(int)) *SimpleList {
	l.selected = handler
	return l
}

// SetDoneFunc sets a function which is called when the user presses the Escape
// key.
func (l *SimpleList) SetDoneFunc(handler func()) *SimpleList {
	l.done = handler
	return l
}

// AddItem adds a new item to the list. An item has a main text which will be
// highlighted when selected. It also has a secondary text which is shown
// underneath the main text (if it is set to visible) but which may remain
// empty.
//
// The shortcut is a key binding. If the specified rune is entered, the item
// is selected immediately. Set to 0 for no binding.
//
// The "selected" callback will be invoked when the user selects the item. You
// may provide nil if no such item is needed or if all events are handled
// through the selected callback set with SetSelectedFunc().
func (l *SimpleList) AddItem(mainText string, selected func()) *SimpleList {
	l.items = append(l.items, &listItem{
		MainText: mainText,
		Selected: selected,
	})
	if len(l.items) == 1 && l.changed != nil {
		l.changed(0)
	}
	return l
}

// GetItemCount returns the number of items in the list.
func (l *SimpleList) GetItemCount() int {
	return len(l.items)
}

// Clear removes all items from the list.
func (l *SimpleList) Clear() *SimpleList {
	l.items = nil
	l.currentItem = 0
	return l
}

// Draw draws this primitive onto the screen.
func (l *SimpleList) Draw(screen tcell.Screen) {
	l.Box.Draw(screen)

	// Determine the dimensions.
	x, y, width, height := l.GetInnerRect()
	bottomLimit := y + height

	// We want to keep the current selection in view. What is our offset?
	var offset int
	if l.currentItem >= height {
		offset = l.currentItem + 1 - height
	}

	// Draw the list items.
	for index, item := range l.items {
		if index < offset {
			continue
		}

		if y >= bottomLimit {
			break
		}

		// Main text.
		tview.Print(screen, item.MainText, x, y, width, tview.AlignLeft, l.mainTextColor)

		// Background color of selected text.
		if index == l.currentItem {
			textWidth := tview.TaggedStringWidth(item.MainText)
			for bx := 0; bx < textWidth && bx < width; bx++ {
				m, c, style, _ := screen.GetContent(x+bx, y)
				fg, _, _ := style.Decompose()
				if fg == l.mainTextColor {
					fg = l.selectedTextColor
				}
				style = style.Background(l.selectedBackgroundColor).Foreground(fg)
				screen.SetContent(x+bx, y, m, c, style)
			}
		}

		y++

		if y >= bottomLimit {
			break
		}
	}
}

// InputHandler returns the handler for this primitive.
func (l *SimpleList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		previousItem := l.currentItem

		switch key := event.Key(); key {
		case tcell.KeyTab, tcell.KeyDown, tcell.KeyRight:
			l.currentItem++
		case tcell.KeyBacktab, tcell.KeyUp, tcell.KeyLeft:
			l.currentItem--
		case tcell.KeyHome:
			l.currentItem = 0
		case tcell.KeyEnd:
			l.currentItem = len(l.items) - 1
		case tcell.KeyPgDn:
			l.currentItem += 5
		case tcell.KeyPgUp:
			l.currentItem -= 5
		case tcell.KeyEnter:
			if l.currentItem >= 0 && l.currentItem < len(l.items) {
				item := l.items[l.currentItem]
				if item.Selected != nil {
					item.Selected()
				}
				if l.selected != nil {
					l.selected(l.currentItem)
				}
			}
		case tcell.KeyEscape:
			if l.done != nil {
				l.done()
			}
		}

		if l.currentItem < 0 {
			l.currentItem = len(l.items) - 1
		} else if l.currentItem >= len(l.items) {
			l.currentItem = 0
		}

		if l.currentItem != previousItem && l.currentItem < len(l.items) && l.changed != nil {
			l.changed(l.currentItem)
		}
	})
}
