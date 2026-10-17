package ui

import (
	"path/filepath"
	"testing"

	"github.com/Nicozilabe/baqueano/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// folderTestModel builds the standard test model wired with an empty folder
// store persisted in a temp directory.
func folderTestModel(t *testing.T) Model {
	t.Helper()
	m := createTestModel()
	store, err := config.LoadFolderStoreFrom(filepath.Join(t.TempDir(), "folders.json"))
	if err != nil {
		t.Fatalf("LoadFolderStoreFrom: %v", err)
	}
	m.folderStore = store
	m.expandedFolders = map[string]bool{}
	m.promptInput = textinput.New()
	m.updateTableColumns()
	m.updateTableRows()
	return m
}

func folderIndex(nodes []treeNode, path string) int {
	for i, n := range nodes {
		if n.kind == nodeKindFolder && n.path == path {
			return i
		}
	}
	return -1
}

func hasHost(nodes []treeNode, name string) bool {
	for _, n := range nodes {
		if n.kind == nodeKindHost && n.name == name {
			return true
		}
	}
	return false
}

func TestTreeGroupsHostsIntoFolders(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.SetHostFolder("server1", "prod"); err != nil {
		t.Fatal(err)
	}
	m.updateTableRows()

	idx := folderIndex(m.treeNodes, "prod")
	if idx < 0 {
		t.Fatal("folder 'prod' not found in tree")
	}
	if m.treeNodes[idx].kind != nodeKindFolder || m.treeNodes[idx].depth != 0 {
		t.Errorf("folder node = %+v, want folder at depth 0", m.treeNodes[idx])
	}
	if !hasHost(m.treeNodes, "server1") {
		t.Error("host server1 not shown inside its folder")
	}
	for _, n := range m.treeNodes {
		if n.kind == nodeKindHost && n.name == "server1" && n.depth != 1 {
			t.Errorf("server1 depth = %d, want 1", n.depth)
		}
	}
}

func TestFolderCollapseAndExpand(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.SetHostFolder("server1", "prod"); err != nil {
		t.Fatal(err)
	}
	m.updateTableRows()

	idx := folderIndex(m.treeNodes, "prod")
	m.table.SetCursor(idx)

	nm, _ := m.handleTreeLeft()
	m = nm
	if hasHost(m.treeNodes, "server1") {
		t.Error("server1 still visible after collapsing its folder")
	}
	if folderIndex(m.treeNodes, "prod") < 0 {
		t.Error("folder disappeared after collapse")
	}

	nm, _ = m.handleTreeRight()
	m = nm
	if !hasHost(m.treeNodes, "server1") {
		t.Error("server1 not visible after expanding its folder")
	}
}

func TestSearchForcesFoldersExpanded(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.SetHostFolder("server1", "prod"); err != nil {
		t.Fatal(err)
	}
	// Collapse the folder first: the search must still show it open.
	m.expandedFolders["prod"] = false

	m.searchInput.SetValue("server1")
	m.filteredHosts = m.filterHosts("server1")
	m.updateTableRows()

	idx := folderIndex(m.treeNodes, "prod")
	if idx < 0 {
		t.Fatal("folder 'prod' missing while searching for a host inside it")
	}
	if !m.treeNodes[idx].expanded {
		t.Error("folder not expanded during search")
	}
	if !hasHost(m.treeNodes, "server1") {
		t.Error("matching host hidden during search")
	}
}

func TestSearchHidesEmptyFolders(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.CreateFolder("empty"); err != nil {
		t.Fatal(err)
	}

	m.searchInput.SetValue("server1")
	m.filteredHosts = m.filterHosts("server1")
	m.updateTableRows()

	if folderIndex(m.treeNodes, "empty") >= 0 {
		t.Error("empty folder should not appear during a search")
	}
}

func TestEmptyFoldersShownWithoutSearch(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.CreateFolder("empty"); err != nil {
		t.Fatal(err)
	}
	m.updateTableRows()

	if folderIndex(m.treeNodes, "empty") < 0 {
		t.Error("empty folder should be visible when not searching")
	}
}

func TestSelectedHostSkipsFolders(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.SetHostFolder("server1", "prod"); err != nil {
		t.Fatal(err)
	}
	m.updateTableRows()

	idx := folderIndex(m.treeNodes, "prod")
	m.table.SetCursor(idx)
	if _, ok := m.selectedHost(); ok {
		t.Error("selectedHost returned a host for a folder row")
	}

	m.table.SetCursor(idx + 1)
	host, ok := m.selectedHost()
	if !ok || host.Name != "server1" {
		t.Errorf("selectedHost on host row = (%q, %v), want server1", host.Name, ok)
	}
}

// Typing a space in the search must not be swallowed by the folder toggle.
func TestSpaceStillTypesInSearch(t *testing.T) {
	m, _ := press(t, folderTestModel(t), slash)
	m, _ = press(t, m, typed("a b")...)
	if got := m.searchInput.Value(); got != "a b" {
		t.Errorf("search value = %q, want %q", got, "a b")
	}
}

func TestMoveHostPromptMovesHost(t *testing.T) {
	m := folderTestModel(t)
	m.table.SetCursor(0)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	if m.promptMode != promptMoveHost {
		t.Fatalf("promptMode = %v, want promptMoveHost", m.promptMode)
	}

	m, _ = press(t, m, typed("prod/web")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.promptMode != promptNone {
		t.Error("prompt stayed open after Enter")
	}
	if got := m.folderStore.FolderOf("server1"); got != "prod/web" {
		t.Errorf("FolderOf(server1) = %q, want prod/web", got)
	}
}

func TestNewFolderPromptCreatesFolder(t *testing.T) {
	m, _ := press(t, folderTestModel(t), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	if m.promptMode != promptNewFolder {
		t.Fatalf("promptMode = %v, want promptNewFolder", m.promptMode)
	}

	m, _ = press(t, m, typed("prod/web")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if folderIndex(m.treeNodes, "prod") < 0 || folderIndex(m.treeNodes, "prod/web") < 0 {
		t.Errorf("created folders missing from tree: %v", m.folderStore.AllFolders())
	}
}

func TestPromptEscapeCancels(t *testing.T) {
	m, _ := press(t, folderTestModel(t), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.promptMode != promptNone {
		t.Error("Escape did not close the prompt")
	}
	if len(m.folderStore.AllFolders()) != 0 {
		t.Error("cancelled prompt created a folder")
	}
}

func TestRenderFolderPromptDoesNotPanic(t *testing.T) {
	m, _ := press(t, folderTestModel(t), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	if out := m.renderFolderPrompt(); out == "" {
		t.Error("renderFolderPrompt returned empty output")
	}
}
