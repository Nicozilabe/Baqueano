package ui

import (
	"os"
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

// TestCreateFolderKeepsCursorOnHost guards against the cursor jumping onto the
// freshly created folder: with the cursor stuck on the folder, "M" could no
// longer select the host to organize.
func TestCreateFolderKeepsCursorOnHost(t *testing.T) {
	m := folderTestModel(t)
	m.table.SetCursor(0) // server1

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	m, _ = press(t, m, typed("prod")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	host, ok := m.selectedHost()
	if !ok || host.Name != "server1" {
		t.Fatalf("after creating a folder, selected host = %q (ok=%v), want server1", host.Name, ok)
	}

	// The move shortcut must now be able to target the folder: it opens a
	// picker listing the existing folders so the name is not typed by hand.
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	if m.promptMode != promptMoveHostPicker {
		t.Fatalf("M did not open the folder picker (promptMode = %v)", m.promptMode)
	}
	// Options are [root, prod, new]; step down to "prod" and confirm.
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if got := m.folderStore.FolderOf("server1"); got != "prod" {
		t.Errorf("FolderOf(server1) = %q, want prod", got)
	}
}

// TestLowercaseMoveFallsBackToFolder verifies that "m" (move to another config
// file) opens the folder prompt when there is no other config file to move the
// host to, instead of failing with "no includes found".
func TestLowercaseMoveFallsBackToFolder(t *testing.T) {
	m := folderTestModel(t)
	base := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(base, []byte("Host server1\n  HostName server1.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m.configFile = base
	m.table.SetCursor(0) // server1

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	if m.promptMode != promptMoveHost {
		t.Fatalf("lowercase m should open the folder prompt when there is no other config file, got promptMode = %v (viewMode = %v)", m.promptMode, m.viewMode)
	}
}

// TestMoveHostReplacesFolder verifies that moving a host that already lives in
// a folder replaces its folder instead of appending to the old one.
func TestMoveHostReplacesFolder(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.SetHostFolder("server1", "prod"); err != nil {
		t.Fatal(err)
	}
	if err := m.folderStore.CreateFolder("staging"); err != nil {
		t.Fatal(err)
	}
	m.updateTableRows()
	hostIdx := -1
	for i, n := range m.treeNodes {
		if n.kind == nodeKindHost && n.name == "server1" {
			hostIdx = i
			break
		}
	}
	if hostIdx < 0 {
		t.Fatal("server1 not found in tree")
	}
	m.table.SetCursor(hostIdx)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	if m.promptMode != promptMoveHostPicker {
		t.Fatalf("promptMode = %v, want promptMoveHostPicker", m.promptMode)
	}
	// The picker should be positioned on the host's current folder.
	if got := m.promptOptions[m.promptCursor]; got != "prod" {
		t.Fatalf("picker preselected %q, want prod", got)
	}
	// Options are [root, prod, staging, new]; step down to "staging".
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if got := m.folderStore.FolderOf("server1"); got != "staging" {
		t.Errorf("FolderOf(server1) = %q, want staging", got)
	}
}

// TestMoveHostPickerShowsFolders verifies the picker lists existing folders
// instead of asking for a free-form name.
func TestMoveHostPickerShowsFolders(t *testing.T) {
	m := folderTestModel(t)
	for _, f := range []string{"prod", "staging/web"} {
		if err := m.folderStore.CreateFolder(f); err != nil {
			t.Fatal(err)
		}
	}
	m.updateTableRows()
	hostIdx := -1
	for i, n := range m.treeNodes {
		if n.kind == nodeKindHost && n.name == "server1" {
			hostIdx = i
			break
		}
	}
	if hostIdx < 0 {
		t.Fatal("server1 not found in tree")
	}
	m.table.SetCursor(hostIdx)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	if m.promptMode != promptMoveHostPicker {
		t.Fatalf("promptMode = %v, want promptMoveHostPicker", m.promptMode)
	}

	seen := map[string]bool{}
	for _, opt := range m.promptOptions {
		seen[opt] = true
	}
	if !seen["prod"] || !seen["staging"] || !seen["staging/web"] {
		t.Errorf("picker options %v do not include every existing folder", m.promptOptions)
	}
	if !seen[promptPickerNewFolder] {
		t.Errorf("picker options %v lack the 'new folder' entry", m.promptOptions)
	}

	if out := m.renderFolderPrompt(); out == "" {
		t.Error("picker render returned empty output")
	}
}

// TestMoveHostPickerNewFolder verifies the picker's "new folder" entry drops
// into the free-form path prompt and still moves the host.
func TestMoveHostPickerNewFolder(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.CreateFolder("prod"); err != nil {
		t.Fatal(err)
	}
	m.updateTableRows()
	hostIdx := -1
	for i, n := range m.treeNodes {
		if n.kind == nodeKindHost && n.name == "server1" {
			hostIdx = i
			break
		}
	}
	m.table.SetCursor(hostIdx)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	if m.promptMode != promptMoveHostPicker {
		t.Fatalf("promptMode = %v, want promptMoveHostPicker", m.promptMode)
	}
	// Jump to the last entry, "+ New folder…".
	for m.promptCursor < len(m.promptOptions)-1 {
		m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if got := m.promptOptions[m.promptCursor]; got != promptPickerNewFolder {
		t.Fatalf("last option = %q, want the new-folder entry", got)
	}

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.promptMode != promptMoveHost {
		t.Fatalf("promptMode = %v, want promptMoveHost after choosing new folder", m.promptMode)
	}
	m, _ = press(t, m, typed("team/db")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if got := m.folderStore.FolderOf("server1"); got != "team/db" {
		t.Errorf("FolderOf(server1) = %q, want team/db", got)
	}
}

// TestMoveHostPickerToRoot verifies the picker can pull a host back to the root.
func TestMoveHostPickerToRoot(t *testing.T) {
	m := folderTestModel(t)
	if err := m.folderStore.SetHostFolder("server1", "prod"); err != nil {
		t.Fatal(err)
	}
	m.updateTableRows()
	hostIdx := -1
	for i, n := range m.treeNodes {
		if n.kind == nodeKindHost && n.name == "server1" {
			hostIdx = i
			break
		}
	}
	m.table.SetCursor(hostIdx)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	// Preselected on "prod"; step up to the root entry.
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyUp})
	if got := m.promptOptions[m.promptCursor]; got != "" {
		t.Fatalf("picker preselected %q, want root", got)
	}
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if got := m.folderStore.FolderOf("server1"); got != "" {
		t.Errorf("FolderOf(server1) = %q, want root", got)
	}
}

// TestRenameFolderReplacesName verifies that typing in the rename prompt
// replaces the folder name instead of appending to it.
func TestRenameFolderReplacesName(t *testing.T) {
	m := folderTestModel(t)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	m, _ = press(t, m, typed("prod")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	idx := folderIndex(m.treeNodes, "prod")
	if idx < 0 {
		t.Fatal("folder prod not found after creation")
	}
	m.table.SetCursor(idx)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	if m.promptMode != promptRenameFolder {
		t.Fatalf("promptMode = %v, want promptRenameFolder", m.promptMode)
	}
	if got := m.promptInput.Value(); got != "" {
		t.Fatalf("rename prompt should start empty, got %q", got)
	}

	m, _ = press(t, m, typed("staging")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	folders := m.folderStore.AllFolders()
	if len(folders) != 1 || folders[0] != "staging" {
		t.Errorf("folders = %v, want [staging]", folders)
	}
	if folderIndex(m.treeNodes, "staging") < 0 {
		t.Errorf("renamed folder not shown in tree: %v", folders)
	}
}

// TestRenameFolderEmptyKeepsName verifies that submitting an empty rename keeps
// the existing folder untouched.
func TestRenameFolderEmptyKeepsName(t *testing.T) {
	m := folderTestModel(t)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	m, _ = press(t, m, typed("prod")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	idx := folderIndex(m.treeNodes, "prod")
	if idx < 0 {
		t.Fatal("folder prod not found after creation")
	}
	m.table.SetCursor(idx)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if folders := m.folderStore.AllFolders(); len(folders) != 1 || folders[0] != "prod" {
		t.Errorf("folders = %v, want [prod]", folders)
	}
	if m.promptMode != promptNone {
		t.Error("empty rename should close the prompt")
	}
}

// TestDeleteFolderClampsCursor verifies the cursor stays within range after the
// highlighted folder is removed.
func TestDeleteFolderClampsCursor(t *testing.T) {
	m := folderTestModel(t)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	m, _ = press(t, m, typed("prod")...)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	idx := folderIndex(m.treeNodes, "prod")
	if idx < 0 {
		t.Fatal("folder prod not found after creation")
	}
	m.table.SetCursor(idx)

	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if len(m.folderStore.AllFolders()) != 0 {
		t.Errorf("folder not deleted: %v", m.folderStore.AllFolders())
	}
	if c := m.table.Cursor(); c < 0 || c >= len(m.treeNodes) {
		t.Errorf("cursor %d out of range for %d rows", c, len(m.treeNodes))
	}
}
