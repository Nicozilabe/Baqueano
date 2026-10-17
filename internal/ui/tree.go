package ui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Nicozilabe/baqueano/internal/config"
)

// nodeKind distinguishes the two kinds of rows shown in the host list.
type nodeKind int

const (
	nodeKindHost nodeKind = iota
	nodeKindFolder
)

// treeNode is a single flattened row of the folder tree. The list view keeps a
// slice of these in display order, aligned with the table rows, so a selection
// can be resolved back to a host or a folder.
type treeNode struct {
	kind        nodeKind
	name        string // host alias or folder base name
	path        string // folder path, only for nodeKindFolder
	depth       int    // nesting depth, used for indentation
	host        config.SSHHost
	expanded    bool
	hasChildren bool
	hostCount   int // number of hosts inside the folder (including subfolders)
}

// folderBranch is the intermediate tree used while building the flattened list.
type folderBranch struct {
	path     string
	name     string
	children map[string]*folderBranch
	hosts    []config.SSHHost
}

func newFolderBranch(path string) *folderBranch {
	return &folderBranch{
		path:     path,
		name:     config.FolderBase(path),
		children: map[string]*folderBranch{},
	}
}

// ensureFolder creates the given folder and every ancestor, returning the leaf.
func (b *folderBranch) ensureFolder(path string) *folderBranch {
	current := b
	for _, ancestor := range config.FolderAncestors(path) {
		child, ok := current.children[ancestor]
		if !ok {
			child = newFolderBranch(ancestor)
			current.children[ancestor] = child
		}
		current = child
	}
	return current
}

// countHosts returns how many hosts live in this branch (subfolders included).
func (b *folderBranch) countHosts() int {
	total := len(b.hosts)
	for _, child := range b.children {
		total += child.countHosts()
	}
	return total
}

// isFolderExpanded reports whether a folder is expanded. While a search is
// active every folder is forced open so matches are always visible.
func (m Model) isFolderExpanded(path string, searching bool) bool {
	if searching {
		return true
	}
	if m.expandedFolders == nil {
		return true
	}
	if expanded, ok := m.expandedFolders[path]; ok {
		return expanded
	}
	return true
}

// folderGlyph returns the disclosure indicator for a folder row.
func folderGlyph(expanded bool) string {
	if expanded {
		return "▾"
	}
	return "▸"
}

// buildTreeNodes flattens the visible hosts into display rows. When no folder
// store is configured it degrades to the historical flat list.
func (m Model) buildTreeNodes(hosts []config.SSHHost) []treeNode {
	nodes := make([]treeNode, 0, len(hosts))

	if m.folderStore == nil {
		for _, host := range hosts {
			nodes = append(nodes, treeNode{kind: nodeKindHost, name: host.Name, host: host})
		}
		return nodes
	}

	root := newFolderBranch("")
	searching := m.searchInput.Value() != ""

	// Register explicitly created folders (including empty ones). During a
	// search only folders that lead to a match are shown, so empty folders do
	// not clutter the results.
	if !searching {
		for _, folder := range m.folderStore.AllFolders() {
			root.ensureFolder(folder)
		}
	}

	// Place every host in its folder, creating folders on the way.
	for _, host := range hosts {
		folder := m.folderStore.FolderOf(host.Name)
		if folder == "" {
			root.hosts = append(root.hosts, host)
			continue
		}
		leaf := root.ensureFolder(folder)
		leaf.hosts = append(leaf.hosts, host)
	}

	m.flattenBranch(root, 0, searching, &nodes)
	return nodes
}

// flattenBranch walks a branch depth-first: subfolders first (alphabetically),
// then the hosts directly inside it. Expanded folders recurse; collapsed ones
// only emit their own row.
func (m Model) flattenBranch(branch *folderBranch, depth int, searching bool, nodes *[]treeNode) {
	names := make([]string, 0, len(branch.children))
	for name := range branch.children {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})

	for _, name := range names {
		child := branch.children[name]
		expanded := m.isFolderExpanded(child.path, searching)
		*nodes = append(*nodes, treeNode{
			kind:        nodeKindFolder,
			name:        child.name,
			path:        child.path,
			depth:       depth,
			expanded:    expanded,
			hasChildren: child.countHosts() > 0 || len(child.children) > 0,
			hostCount:   child.countHosts(),
		})
		if expanded {
			m.flattenBranch(child, depth+1, searching, nodes)
		}
	}

	for _, host := range branch.hosts {
		*nodes = append(*nodes, treeNode{
			kind:  nodeKindHost,
			name:  host.Name,
			depth: depth,
			host:  host,
		})
	}
}

// folderDisplayLabel builds the "Name" column text for a folder row.
func folderDisplayLabel(node treeNode) string {
	indent := strings.Repeat("  ", node.depth)
	label := indent + folderGlyph(node.expanded) + " " + node.name + "/"
	if node.hostCount > 0 {
		label += " (" + strconv.Itoa(node.hostCount) + ")"
	}
	return label
}

// hostDisplayLabel builds the "Name" column text for a host row.
func hostDisplayLabel(indent int, statusIndicator, name string) string {
	return strings.Repeat("  ", indent) + statusIndicator + " " + name
}
