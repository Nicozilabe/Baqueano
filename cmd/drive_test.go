package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Gu1llaum-3/sshm/internal/drive"

	"github.com/spf13/cobra"
)

// isolateDriveConfig points the sshm config dir at a throwaway temp dir so
// credential tests never touch the real user config.
func isolateDriveConfig(t *testing.T) {
	t.Helper()
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSHM_GOOGLE_CLIENT_ID", "")
	t.Setenv("SSHM_GOOGLE_CLIENT_SECRET", "")
}

// newPromptCmd returns a command whose stdin/stdout are wired to the given
// input and an in-memory output buffer.
func newPromptCmd(in string, out *bytes.Buffer) *cobra.Command {
	c := &cobra.Command{}
	c.SetIn(strings.NewReader(in))
	c.SetOut(out)
	c.SetErr(out)
	return c
}

func TestReadLineStripsCRLF(t *testing.T) {
	got, err := readLine(strings.NewReader("hello\r\n"))
	if err != nil {
		t.Fatalf("readLine: %v", err)
	}
	if got != "hello" {
		t.Errorf("readLine = %q, want %q", got, "hello")
	}
}

func TestReadLineEOFWithoutNewline(t *testing.T) {
	got, err := readLine(strings.NewReader("last"))
	if err != nil {
		t.Fatalf("readLine: %v", err)
	}
	if got != "last" {
		t.Errorf("readLine = %q, want %q", got, "last")
	}
}

func TestPromptCredentialsReadsIDAndSecret(t *testing.T) {
	var out bytes.Buffer
	cmd := newPromptCmd("cid-123\nsecret-456\n", &out)

	creds, err := promptCredentials(cmd)
	if err != nil {
		t.Fatalf("promptCredentials: %v", err)
	}
	if creds.ClientID != "cid-123" || creds.ClientSecret != "secret-456" {
		t.Errorf("credentials = %+v, want id=cid-123 secret=secret-456", creds)
	}
	if !strings.Contains(out.String(), "Client ID:") || !strings.Contains(out.String(), "Client Secret:") {
		t.Errorf("prompt output missing field labels:\n%s", out.String())
	}
}

func TestResolveLoginCredentialsUsesFlagsAndSaves(t *testing.T) {
	isolateDriveConfig(t)

	origID, origSecret := driveLoginClientID, driveLoginClientSecret
	t.Cleanup(func() { driveLoginClientID, driveLoginClientSecret = origID, origSecret })

	driveLoginClientID = "flag-id"
	driveLoginClientSecret = "flag-secret"

	var out bytes.Buffer
	creds, err := resolveLoginCredentials(newPromptCmd("", &out))
	if err != nil {
		t.Fatalf("resolveLoginCredentials: %v", err)
	}
	if creds.ClientID != "flag-id" || creds.ClientSecret != "flag-secret" {
		t.Errorf("credentials = %+v, want the flag values", creds)
	}

	saved, err := drive.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials after save: %v", err)
	}
	if saved != creds {
		t.Errorf("saved credentials = %+v, want %+v", saved, creds)
	}
}

func TestResolveLoginCredentialsPromptsWhenMissing(t *testing.T) {
	isolateDriveConfig(t)

	origID, origSecret := driveLoginClientID, driveLoginClientSecret
	t.Cleanup(func() { driveLoginClientID, driveLoginClientSecret = origID, origSecret })
	driveLoginClientID, driveLoginClientSecret = "", ""

	var out bytes.Buffer
	creds, err := resolveLoginCredentials(newPromptCmd("p-id\np-secret\n", &out))
	if err != nil {
		t.Fatalf("resolveLoginCredentials: %v", err)
	}
	if creds.ClientID != "p-id" || creds.ClientSecret != "p-secret" {
		t.Errorf("credentials = %+v, want the prompted values", creds)
	}

	saved, err := drive.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials after prompt save: %v", err)
	}
	if saved.ClientID != "p-id" {
		t.Errorf("saved client id = %q, want p-id", saved.ClientID)
	}
}

func TestDriveLoginCmdRegistered(t *testing.T) {
	found := false
	for _, c := range driveCmd.Commands() {
		if c.Name() == "login" {
			found = true
		}
	}
	if !found {
		t.Fatal("drive login command is not registered")
	}
}
