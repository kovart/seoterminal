// This is a changed clone of Tview Component (primitive)

package main

import "strings"

// ClusterNode represents one node in a tree view.
type ClusterNode struct {
	*Cluster
	Name       string
	IsExpanded bool
	Parent     *ClusterNode // The Parent node (nil for the root).
	Neighbors  []*ClusterNode
	level      int
	children   []*ClusterNode
}

// NewClusterNode returns a new tree node.
func NewClusterNode(name string, node *Cluster, isExpanded bool, parent *ClusterNode) *ClusterNode {
	return &ClusterNode{
		Name:       name,
		IsExpanded: isExpanded,
		Parent:     parent,
		Cluster:    node,
	}
}

// Walk traverses this node's subtree in depth-first, pre-order (NLR) order and
// calls the provided callback function on each traversed node (which includes
// this node) with the traversed node and its parent node (nil for this node).
// The callback returns whether traversal should continue with the traversed
// node's child nodes (true) or not recurse any deeper (false).
func (n *ClusterNode) Walk(callback func(node, parent *ClusterNode) bool) *ClusterNode {
	//n.Parent = nil
	nodes := []*ClusterNode{n}
	for len(nodes) > 0 {
		// Pop the top node and process it.
		node := nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
		if !callback(node, node.Parent) {
			// Don't add any children.
			continue
		}

		// Add children in reverse order.
		for index := len(node.children) - 1; index >= 0; index-- {
			node.children[index].Parent = node
			nodes = append(nodes, node.children[index])
		}
	}
	return n
}

// SetChildren sets this node's child nodes.
func (n *ClusterNode) SetChildren(childNodes []*ClusterNode) *ClusterNode {
	n.children = childNodes
	return n
}

// ClearChildren removes all child nodes from this node.
func (n *ClusterNode) ClearChildren() *ClusterNode {
	n.children = nil
	return n
}

// AddChild adds a new child node to this node.
func (n *ClusterNode) AddChild(node *ClusterNode) *ClusterNode {
	n.children = append(n.children, node)
	return n
}

// SetExpanded sets whether or not this node's child nodes should be displayed.
func (n *ClusterNode) SetExpanded(expanded bool) *ClusterNode {
	n.IsExpanded = expanded
	return n
}

// Expand makes the child nodes of this node appear.
func (n *ClusterNode) Expand() *ClusterNode {
	n.IsExpanded = true
	return n
}

// Collapse makes the child nodes of this node disappear.
func (n *ClusterNode) Collapse() *ClusterNode {
	n.IsExpanded = false
	return n
}

// ExpandAll expands this node and all descendent nodes.
func (n *ClusterNode) ExpandAll() *ClusterNode {
	n.Walk(func(node, parent *ClusterNode) bool {
		node.IsExpanded = true
		return true
	})
	return n
}

// CollapseAll collapses this node and all descendent nodes.
func (n *ClusterNode) CollapseAll() *ClusterNode {
	n.Walk(func(node, parent *ClusterNode) bool {
		n.IsExpanded = false
		return true
	})
	return n
}

// SetName sets the node's Name which is displayed.
func (n *ClusterNode) SetName(text string) *ClusterNode {
	n.Name = text
	return n
}

// Parameter 'existedCluster' can be nil
func (n *ClusterNode) GenerateChildren(existedClusters map[string]*Cluster) (ret []*ClusterNode) {
	nodeMap := n.Cluster.GenerateSubClusters(existedClusters, 3)
	for key, node := range nodeMap {
		cluster := NewClusterNode(key, node, false, n)
		//if n.State == NodeRemoved {
		//	cluster.State = n.State
		//}
		ret = append(ret, cluster)
	}
	return
}

func (n *ClusterNode) GetFullName() string {
	s := n.GetClusterNames()
	return strings.Join(s, " ")
}

func (n *ClusterNode) GetClusterNames() []string {
	names := []string{n.Name}
	node := n
	for node.Parent != nil && node.Parent.Hash != "" {
		names = append(names, node.Parent.Name)
		node = node.Parent
	}
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return names
}
