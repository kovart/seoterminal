// This is a changed clone of Tview Component (primitive)

package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"strings"
)

// Tree navigation events.
const (
	treeNone int = iota
	treeHome
	treeEnd
	treeUp
	treeDown
)

type ClusterTreeView struct {
	*tview.Box

	// The root node.
	root *ClusterNode

	// The currently selectedCallback node or nil if no node is selectedCallback.
	currentNode *ClusterNode

	// The movement to be performed during the call to Draw(), one of the
	// constants defined above.
	movement int

	// Vertical scroll offset.
	offsetY int

	isRootVisible bool

	// An optional function which is called when the user has navigated to a new
	// tree node.
	navigatedCallback func(node *ClusterNode)

	// An optional function which is called when a tree item was selectedCallback.
	selectedCallback func(node *ClusterNode)

	// An optional function which is called when some action was applied to tree item
	actionCallback func(node *ClusterNode)

	controlCallback func(node *ClusterNode, key *tcell.EventKey)

	// The visible nodes, top-down, as set by process().
	nodes []*ClusterNode
}

// NewTreeView returns a new tree view.
func NewKeywordTreeView() *ClusterTreeView {
	return &ClusterTreeView{
		Box:           tview.NewBox(),
		isRootVisible: true,
	}
}

// SetRoot sets the root node of the tree.
func (t *ClusterTreeView) SetRoot(root *ClusterNode) *ClusterTreeView {
	t.root = root
	return t
}

// GetRoot returns the root node of the tree. If no such node was previously
// set, nil is returned.
func (t *ClusterTreeView) GetRoot() *ClusterNode {
	return t.root
}

// SetCurrentNode sets the currently selectedCallback node. Provide nil to clear all
// selections. Selected nodes must be visible and selectable, or else the
// selection will be navigatedCallback to the top-most selectable and visible node.
//
// This function does NOT trigger the "navigatedCallback" callback.
func (t *ClusterTreeView) SetCurrentNode(node *ClusterNode) *ClusterTreeView {
	t.currentNode = node
	return t
}

// GetCurrentNode returns the currently selectedCallback node or nil of no node is
// currently selectedCallback.
func (t *ClusterTreeView) GetCurrentNode() *ClusterNode {
	return t.currentNode
}

// SetTopLevel sets the first tree level that is visible with 0 referring to the
// root, 1 to the root's child nodes, and so on. Nodes above the top level are
// not displayed.
func (t *ClusterTreeView) SetIsRootVisible(show bool) *ClusterTreeView {
	t.isRootVisible = show
	return t
}

// SetNavigatedFunc sets the function which is called when the user navigates to
// a new tree node.
func (t *ClusterTreeView) SetNavigatedFunc(handler func(node *ClusterNode)) *ClusterTreeView {
	t.navigatedCallback = handler
	return t
}

// SetSelectedFunc sets the function which is called when the user selects a
// node by pressing Enter on the current selection.
func (t *ClusterTreeView) SetSelectedFunc(handler func(node *ClusterNode)) *ClusterTreeView {
	t.selectedCallback = handler
	return t
}

func (t *ClusterTreeView) SetActionFunc(handler func(node *ClusterNode)) *ClusterTreeView {
	t.actionCallback = handler
	return t
}

func (t *ClusterTreeView) SetControlFunc(handler func(node *ClusterNode, key *tcell.EventKey)) *ClusterTreeView {
	t.controlCallback = handler
	return t
}

// process builds the visible tree, populates the "nodes" slice, and processes
// pending selection actions.
func (t *ClusterTreeView) process() {
	_, _, _, height := t.GetInnerRect()

	t.nodes = nil
	selectedIndex := -1
	t.root.Walk(func(node, parent *ClusterNode) bool {
		// Set node attributes.
		node.Parent = parent
		if parent == nil {
			node.level = 0
		} else {
			node.level = parent.level + 1
		}
		if node == t.currentNode {
			selectedIndex = len(t.nodes)
		}

		// Add and recurse (if desired).
		if t.isRootVisible || node.level > 0 {
			t.nodes = append(t.nodes, node)
		}
		return node.IsExpanded
	})

	// Set neighbors
	for i := 0; i < len(t.nodes); i++ {
		node := t.nodes[i]
		var neighbors []*ClusterNode
		if i-1 >= 0 {
			neighbors = append(neighbors, t.nodes[i-1])
		}
		if i+1 < len(t.nodes) {
			neighbors = append(neighbors, t.nodes[i+1])
		}
		node.Neighbors = neighbors
	}

	// Process selection. (Also trigger events if necessary.)
	if selectedIndex >= 0 {
		// Move the selection.
		newSelectedIndex := selectedIndex
	MovementSwitch:
		switch t.movement {
		case treeUp:
			for newSelectedIndex > 0 {
				newSelectedIndex--
				break MovementSwitch
			}
			newSelectedIndex = selectedIndex
		case treeDown:
			for newSelectedIndex < len(t.nodes)-1 {
				newSelectedIndex++
				break MovementSwitch
			}
			newSelectedIndex = selectedIndex
		case treeHome:
			for newSelectedIndex = 0; newSelectedIndex < len(t.nodes); newSelectedIndex++ {
				break MovementSwitch
			}
			newSelectedIndex = selectedIndex
		case treeEnd:
			for newSelectedIndex = len(t.nodes) - 1; newSelectedIndex >= 0; newSelectedIndex-- {
				break MovementSwitch
			}
			newSelectedIndex = selectedIndex
		}

		t.currentNode = t.nodes[newSelectedIndex]
		if newSelectedIndex != selectedIndex {
			t.movement = treeNone
			if t.navigatedCallback != nil {
				t.navigatedCallback(t.currentNode)
			}
		}
		selectedIndex = newSelectedIndex

		// Move selection into viewport.
		if selectedIndex-t.offsetY >= height {
			t.offsetY = selectedIndex - height + 1
		}
		if selectedIndex < t.offsetY {
			t.offsetY = selectedIndex
		}
	} else {
		// If selection is not visible or selectable, select the first candidate.
		if t.currentNode != nil {
			for index, node := range t.nodes {
				selectedIndex = index
				t.currentNode = node
				break
			}
		}
		if selectedIndex < 0 {
			t.currentNode = nil
		}
	}

}

// Draw draws this primitive onto the screen.
func (t *ClusterTreeView) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)
	if t.root == nil {
		return
	}

	// Build the tree if necessary.
	if t.nodes == nil {
		t.process()
	}
	defer func() {
		t.nodes = nil // Rebuild during next call to Draw()
	}()

	// Scroll the tree.
	x, y, width, height := t.GetInnerRect()
	switch t.movement {
	case treeUp:
		t.offsetY--
	case treeDown:
		t.offsetY++
	case treeHome:
		t.offsetY = 0
	case treeEnd:
		t.offsetY = len(t.nodes)
	}
	t.movement = treeNone

	// Fix invalid offsets.
	if t.offsetY >= len(t.nodes)-height {
		t.offsetY = len(t.nodes) - height
	}
	if t.offsetY < 0 {
		t.offsetY = 0
	}

	posY := y
	for index, node := range t.nodes {
		// Skip invisible parts.
		if posY >= y+height {
			break
		}
		if index < t.offsetY {
			continue
		}

		var color string
		if t.currentNode == node {
			color = "[black:white]"
		}
		var line string
		text := node.Name
		indent := strings.Repeat(" ", node.level*3)
		line = fmt.Sprintf(` %s%s[•] %s `, color, indent, text)
		tview.Print(screen, line, x, posY, width, tview.AlignLeft, tcell.ColorWhite)

		// Advance.
		posY++
	}
}

// InputHandler returns the handler for this primitive.
func (t *ClusterTreeView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Because the tree is flattened into a list only at drawing time, we also
		// postpone the (selection) movement to drawing time.
		switch key := event.Key(); key {
		case tcell.KeyDown, tcell.KeyRight:
			t.movement = treeDown
		case tcell.KeyUp, tcell.KeyLeft:
			t.movement = treeUp
		case tcell.KeyHome:
			t.movement = treeHome
		case tcell.KeyEnd:
			t.movement = treeEnd
		case tcell.KeyEnter:
			if t.currentNode != nil {
				if t.selectedCallback != nil {
					t.selectedCallback(t.currentNode)
				}
			} else {
				t.SetCurrentNode(t.GetRoot())
			}
		}

		if event.Modifiers()&tcell.ModCtrl != 0 ||
			event.Modifiers()&tcell.ModShift != 0 ||
			event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 ||
			event.Key() == tcell.KeyCtrlBackslash ||
			event.Rune() == '-' || event.Rune() == '+' || event.Rune() == '/' {
			if t.controlCallback != nil {
				t.controlCallback(t.currentNode, event)
			}
		}

		t.process()
	})
}
