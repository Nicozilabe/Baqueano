package ui

import (
	"github.com/Nicozilabe/baqueano/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// selectedNode returns the tree node under the cursor, if any.
func (m Model) selectedNode() (treeNode, bool) {
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.treeNodes) {
		return m.treeNodes[cursor], true
	}
	return treeNode{}, false
}

// selectedHost returns the host under the cursor. It reports false when the
// cursor is on a folder (or on nothing).
func (m Model) selectedHost() (config.SSHHost, bool) {
	if node, ok := m.selectedNode(); ok {
		if node.kind == nodeKindHost {
			return node.host, true
		}
		return config.SSHHost{}, false
	}
	// Fallback for callers that have not (re)built the tree rows yet.
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filteredHosts) {
		return m.filteredHosts[cursor], true
	}
	return config.SSHHost{}, false
}

// selectedFolder returns the folder under the cursor, if any.
func (m Model) selectedFolder() (string, bool) {
	if node, ok := m.selectedNode(); ok && node.kind == nodeKindFolder {
		return node.path, true
	}
	return "", false
}

// nodeKey identifies a table row across rebuilds (folders by path, hosts by
// name) so the cursor can stay on the same logical row after the tree changes.
func nodeKey(node treeNode) string {
	if node.kind == nodeKindFolder {
		return "F:" + node.path
	}
	return "H:" + node.name
}

// cursorKey returns the identity of the row under the cursor ("" when none).
func (m Model) cursorKey() string {
	if node, ok := m.selectedNode(); ok {
		return nodeKey(node)
	}
	return ""
}

// restoreCursorTo puts the cursor back on the row with the given key. When that
// row is gone (e.g. a deleted folder) the cursor is clamped into range instead.
func (m *Model) restoreCursorTo(key string) {
	if key != "" {
		for i, node := range m.treeNodes {
			if nodeKey(node) == key {
				m.table.SetCursor(i)
				return
			}
		}
	}
	m.clampCursor()
}

// clampCursor keeps the cursor inside the current row range.
func (m *Model) clampCursor() {
	if len(m.treeNodes) == 0 {
		return
	}
	if cursor := m.table.Cursor(); cursor >= len(m.treeNodes) {
		m.table.SetCursor(len(m.treeNodes) - 1)
	} else if cursor < 0 {
		m.table.SetCursor(0)
	}
}

// setFolderExpanded stores the expanded state of a folder. Ancestor folders are
// expanded too, so revealing a nested folder never leaves a collapsed parent.
func (m *Model) setFolderExpanded(path string, expanded bool) {
	if m.expandedFolders == nil {
		m.expandedFolders = map[string]bool{}
	}
	for _, ancestor := range config.FolderAncestors(path) {
		m.expandedFolders[ancestor] = expanded
	}
}

// toggleFolder flips the expanded state of a folder.
func (m *Model) toggleFolder(path string) {
	if m.searchInput.Value() != "" {
		// Search always shows folders expanded; toggling is a no-op there.
		return
	}
	current := m.isFolderExpanded(path, false)
	m.setFolderExpanded(path, !current)
}

// activateFolder is what Enter/Space do on a folder: move into it when a search
// keeps it open, otherwise collapse/expand it.
func (m *Model) activateFolder(node treeNode) {
	cursor := m.table.Cursor()
	if m.searchInput.Value() != "" {
		if node.hasChildren {
			m.table.SetCursor(cursor + 1)
		}
		return
	}
	m.toggleFolder(node.path)
	m.updateTableRows()
	// Keep the cursor on the folder that was just toggled.
	if cursor < len(m.treeNodes) {
		m.table.SetCursor(cursor)
	}
}

// handleTreeRight is the "enter the folder with the right arrow" action:
// expand a collapsed folder, or step into the first child of an open one.
func (m *Model) handleTreeRight() (Model, tea.Cmd) {
	node, ok := m.selectedNode()
	if !ok || node.kind != nodeKindFolder {
		return *m, nil
	}

	if !m.isFolderExpanded(node.path, m.searchInput.Value() != "") {
		m.setFolderExpanded(node.path, true)
		m.updateTableRows()
		return *m, nil
	}
	if node.hasChildren {
		m.table.SetCursor(m.table.Cursor() + 1)
	}
	return *m, nil
}

// handleTreeLeft collapses an open folder, or walks back up to the parent
// folder from a collapsed folder or a host.
func (m *Model) handleTreeLeft() (Model, tea.Cmd) {
	node, ok := m.selectedNode()
	if !ok {
		return *m, nil
	}

	if node.kind == nodeKindFolder && m.isFolderExpanded(node.path, m.searchInput.Value() != "") {
		m.setFolderExpanded(node.path, false)
		m.updateTableRows()
		return *m, nil
	}

	m.moveCursorToParent()
	return *m, nil
}

// moveCursorToParent puts the cursor on the nearest folder above the current
// row (the folder that contains the selected node).
func (m *Model) moveCursorToParent() {
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.treeNodes) {
		return
	}
	depth := m.treeNodes[cursor].depth
	for i := cursor - 1; i >= 0; i-- {
		if m.treeNodes[i].kind == nodeKindFolder && m.treeNodes[i].depth == depth-1 {
			m.table.SetCursor(i)
			return
		}
	}
}

// ensureFolderStore lazily (re)loads the folder store if it was not available
// at startup.
func (m *Model) ensureFolderStore() *config.FolderStore {
	if m.folderStore == nil {
		if store, err := config.LoadFolderStore(); err == nil {
			m.folderStore = store
		}
	}
	return m.folderStore
}

// showFolderError surfaces a short-lived error banner in the list view.
func (m *Model) showFolderError(message string) {
	m.errorMessage = message
	m.showingError = true
}

// openNewFolderPrompt asks for the path of a folder to create. When a folder is
// selected the prompt is prefilled with it so subfolders are one keystroke away.
func (m *Model) openNewFolderPrompt() (Model, tea.Cmd) {
	if m.ensureFolderStore() == nil {
		m.showFolderError("Folder organization is unavailable")
		return *m, nil
	}

	prefill := ""
	if folder, ok := m.selectedFolder(); ok {
		prefill = folder + "/"
	}
	m.openPrompt(promptNewFolder, "New folder", "prod, prod/web, ...", prefill)
	return *m, textinput.Blink
}

// openMoveHostPrompt asks which folder the selected host should move into.
func (m *Model) openMoveHostPrompt() (Model, tea.Cmd) {
	store := m.ensureFolderStore()
	if store == nil {
		m.showFolderError("Folder organization is unavailable")
		return *m, nil
	}

	host, ok := m.selectedHost()
	if !ok {
		m.showFolderError("Select a host to move into a folder")
		return *m, nil
	}

	m.promptHost = host.Name
	m.openPrompt(promptMoveHost, "Move "+host.Name+" to folder", "folder name (empty = root)", store.FolderOf(host.Name))
	return *m, textinput.Blink
}

// openRenameFolderPrompt asks for the new path of the selected folder.
func (m *Model) openRenameFolderPrompt() (Model, tea.Cmd) {
	if m.ensureFolderStore() == nil {
		m.showFolderError("Folder organization is unavailable")
		return *m, nil
	}

	folder, ok := m.selectedFolder()
	if !ok {
		m.showFolderError("Select a folder to rename")
		return *m, nil
	}

	m.promptFolder = folder
	m.openPrompt(promptRenameFolder, "Rename folder", "new folder name", folder)
	return *m, textinput.Blink
}

// openDeleteFolderPrompt asks for confirmation before deleting a folder.
func (m *Model) openDeleteFolderPrompt() (Model, tea.Cmd) {
	if m.ensureFolderStore() == nil {
		m.showFolderError("Folder organization is unavailable")
		return *m, nil
	}

	folder, ok := m.selectedFolder()
	if !ok {
		m.showFolderError("Select a folder to delete")
		return *m, nil
	}

	m.promptFolder = folder
	m.promptMode = promptDeleteFolder
	m.promptError = ""
	m.table.Blur()
	return *m, nil
}

// openPrompt sets up the shared inline text prompt.
func (m *Model) openPrompt(mode promptMode, title, placeholder, value string) {
	ti := m.promptInput
	ti.SetValue(value)
	ti.Placeholder = placeholder
	ti.CursorEnd()
	ti.Focus()
	m.promptInput = ti
	m.promptTitle = title
	m.promptMode = mode
	m.promptError = ""
	m.table.Blur()
}

// closePrompt dismisses the inline prompt and returns focus to the table.
func (m *Model) closePrompt() {
	m.promptMode = promptNone
	m.promptHost = ""
	m.promptFolder = ""
	m.promptError = ""
	m.promptInput.SetValue("")
	m.promptInput.Blur()
	m.table.Focus()
}

// handlePromptKeys routes keys while an inline prompt is open.
func (m Model) handlePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closePrompt()
		return m, nil
	case "enter":
		return m.submitPrompt()
	}

	if m.promptMode == promptDeleteFolder {
		if msg.String() == "y" {
			return m.submitPrompt()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.promptInput, cmd = m.promptInput.Update(msg)
	m.promptError = ""
	return m, cmd
}

// submitPrompt applies the action of the currently open prompt.
func (m Model) submitPrompt() (tea.Model, tea.Cmd) {
	store := m.folderStore
	if store == nil {
		m.closePrompt()
		m.showFolderError("Folder organization is unavailable")
		return m, nil
	}

	// Remember the row the cursor was on so the selection survives the rebuild.
	key := m.cursorKey()

	switch m.promptMode {
	case promptNewFolder:
		path := config.NormalizeFolderPath(m.promptInput.Value())
		if path == "" {
			m.promptError = "Enter a folder name"
			return m, nil
		}
		if err := store.CreateFolder(path); err != nil {
			m.promptError = err.Error()
			return m, nil
		}
		m.ensureFolderVisible(path)

	case promptMoveHost:
		folder := config.NormalizeFolderPath(m.promptInput.Value())
		if err := store.SetHostFolder(m.promptHost, folder); err != nil {
			m.promptError = err.Error()
			return m, nil
		}
		if folder != "" {
			m.ensureFolderVisible(folder)
		}
		if m.promptHost != "" {
			key = "H:" + m.promptHost
		}

	case promptRenameFolder:
		newPath := config.NormalizeFolderPath(m.promptInput.Value())
		if newPath == "" {
			m.promptError = "Enter a folder name"
			return m, nil
		}
		if err := store.RenameFolder(m.promptFolder, newPath); err != nil {
			m.promptError = err.Error()
			return m, nil
		}
		m.ensureFolderVisible(newPath)
		key = "F:" + newPath

	case promptDeleteFolder:
		if err := store.DeleteFolder(m.promptFolder); err != nil {
			m.promptError = err.Error()
			return m, nil
		}
		// The folder is gone: fall back to clamping the cursor.
		key = ""
	}

	m.closePrompt()
	m.updateTableRows()
	m.restoreCursorTo(key)
	return m, nil
}

// ensureFolderVisible expands a folder and its ancestors so a freshly created
// or newly populated folder is shown open.
func (m *Model) ensureFolderVisible(path string) {
	m.setFolderExpanded(path, true)
}

// promptInputWidth caps the input width to the terminal size.
func (m Model) promptInputWidth() int {
	w := m.width - 20
	if w < 20 {
		w = 20
	}
	if w > 60 {
		w = 60
	}
	return w
}
