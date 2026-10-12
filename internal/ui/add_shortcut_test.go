package ui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// AC: pressing "a" in list view opens the add-host form
func TestKeyACode(t *testing.T) {
	m := createTestModel()
	m.configFile = filepath.Join(t.TempDir(), "config")

	m, cmd := press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		t.Error("press a returned no command, want textinput.Blink")
	}
	if m.viewMode != ViewAdd {
		t.Errorf("viewMode = %v, want ViewAdd", m.viewMode)
	}
	if m.addForm == nil {
		t.Fatal("addForm is nil, want the add form")
	}
	// The add form must render a usable view
	if v := m.View(); v == "" {
		t.Error("View() is empty while in add mode")
	}
}