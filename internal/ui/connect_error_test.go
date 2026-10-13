package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/Nicozilabe/baqueano/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// connectTestModel returns a model whose only host is a local, always-refused
// endpoint (127.0.0.1 on a closed port), so the preflight fails fast and
// deterministically.
func connectTestModel() Model {
	m := createTestModel()
	m.hosts = []config.SSHHost{{Name: "bad", Hostname: "127.0.0.1", Port: "1"}}
	m.filteredHosts = m.hosts
	m.updateTableRows()
	return m
}

// Pressing Enter kicks off a connection: the host is marked as connecting and
// a command is returned instead of launching ssh directly.
func TestStartConnectStartsPendingState(t *testing.T) {
	m := connectTestModel()
	m, cmd := m.startConnect("bad")
	if cmd == nil {
		t.Fatal("startConnect returned no command")
	}
	if m.connectingHost != "bad" {
		t.Errorf("connectingHost = %q, want %q", m.connectingHost, "bad")
	}
}

// A failed preflight surfaces a persistent error banner instead of quitting.
func TestPreflightFailureShowsPersistentError(t *testing.T) {
	m := connectTestModel()
	preflight := m.preflightAndConnect("bad")
	msg := preflight()
	pf, ok := msg.(connectPreflightMsg)
	if !ok {
		t.Fatalf("preflight returned %T, want connectPreflightMsg", msg)
	}
	if pf.err == nil {
		t.Fatal("preflight to 127.0.0.1:1 succeeded, want connection error")
	}

	m.connectingHost = "bad"
	tm, cmd := m.Update(pf)
	m = tm.(Model)
	if cmd != nil {
		t.Errorf("Update returned a command, want nil (no auto-clear timer)")
	}
	if !m.showingError {
		t.Error("showingError = false, want true")
	}
	if m.connectingHost != "" {
		t.Errorf("connectingHost = %q, want empty", m.connectingHost)
	}
	if !strings.Contains(m.errorMessage, "Connection failed with bad") {
		t.Errorf("errorMessage = %q, want it to name the host", m.errorMessage)
	}
	if !strings.Contains(m.errorMessage, "Press ENTER, ESC or SPACE to continue") {
		t.Errorf("errorMessage = %q, want the dismiss hint", m.errorMessage)
	}
}

// A successful preflight hands control to the real ssh command.
func TestPreflightSuccessLaunchesConnector(t *testing.T) {
	m := connectTestModel()
	m.connectingHost = "bad"
	tm, cmd := m.Update(connectPreflightMsg{hostName: "bad"})
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("Update returned no command, want the ssh command")
	}
	if m.connectingHost != "" {
		t.Errorf("connectingHost = %q, want empty", m.connectingHost)
	}
}

// An interactive ssh session that ends in error shows a persistent banner and
// returns to the list view even when it ran from another view (port forward).
func TestExecFinishedErrorIsPersistent(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewPortForward
	m.portForwardForm = NewPortForwardForm("server1", m.styles, m.width, m.height, m.configFile, m.historyManager)

	tm, cmd := m.Update(execFinishedMsg{host: "server1", err: errors.New("boom")})
	m = tm.(Model)
	if cmd != nil {
		t.Error("Update returned a command, want nil (persistent banner)")
	}
	if !m.showingError {
		t.Error("showingError = false, want true")
	}
	if m.viewMode != ViewList {
		t.Errorf("viewMode = %v, want ViewList", m.viewMode)
	}
	if m.portForwardForm != nil {
		t.Error("portForwardForm not cleared")
	}
	if !strings.Contains(m.errorMessage, "Connection failed with server1") ||
		!strings.Contains(m.errorMessage, "boom") {
		t.Errorf("errorMessage = %q, want host, reason and dismiss hint", m.errorMessage)
	}
}

// A clean ssh session (exit 0) still quits the app.
func TestExecFinishedSuccessQuits(t *testing.T) {
	m := connectTestModel()
	tm, cmd := m.Update(execFinishedMsg{host: "bad"})
	_ = tm.(Model)
	if !quits(cmd) {
		t.Error("successful session did not quit")
	}
}

// The error banner stays until the user presses a dismiss key.
func TestConnectionErrorDismissedByKey(t *testing.T) {
	for _, key := range []tea.KeyMsg{
		tea.KeyMsg{Type: tea.KeyEnter},
		tea.KeyMsg{Type: tea.KeyEsc},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
	} {
		tm, _ := createTestModel().Update(connectPreflightMsg{hostName: "bad", err: errors.New("refused")})
		m := tm.(Model)
		if !m.showingError {
			t.Fatalf("setup before %q: expected error banner", key)
		}
		m, _ = press(t, m, key)
		if m.showingError || m.errorMessage != "" {
			t.Errorf("key %q did not dismiss the banner (showingError=%v, msg=%q)",
				key.String(), m.showingError, m.errorMessage)
		}
	}
}

// While the error banner is up, navigation keys are ignored.
func TestKeysIgnoredWhileErrorShown(t *testing.T) {
	tm, _ := createTestModel().Update(connectPreflightMsg{hostName: "bad", err: errors.New("refused")})
	m := tm.(Model)
	if !m.showingError {
		t.Fatal("setup: expected error banner")
	}

	m, cmd := press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	if m.searchMode {
		t.Error("slash entered search while the error banner was up")
	}
	if cmd != nil {
		t.Error("navigation returned a command while the banner was up")
	}
	if !m.showingError {
		t.Error("banner cleared by a non-dismiss key")
	}
}

// The connecting spinner advances on each tick and stops once the attempt is over.
func TestSpinnerTicksWhileConnecting(t *testing.T) {
	m := connectTestModel()
	m, _ = m.startConnect("bad")

	tm, cmd := m.Update(spinnerTickMsg{})
	m = tm.(Model)
	if m.spinnerFrame != 1 {
		t.Errorf("spinnerFrame after first tick = %d, want 1", m.spinnerFrame)
	}
	if cmd == nil {
		t.Error("tick returned no command, want the next tick")
	}

	tm, _ = m.Update(connectPreflightMsg{hostName: "bad", err: errors.New("refused")})
	m = tm.(Model)
	if m.connectingHost != "" {
		t.Fatalf("connectingHost = %q, want empty", m.connectingHost)
	}
	frame := m.spinnerFrame
	tm, cmd = m.Update(spinnerTickMsg{})
	m = tm.(Model)
	if cmd != nil {
		t.Error("tick returned a command while not connecting")
	}
	if m.spinnerFrame != frame {
		t.Error("spinnerFrame advanced while not connecting")
	}
}