package drive

import (
	"errors"
	"os"
	"runtime"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// isolateConfigDir redirects the sshm config dir (and the OS config dir on
// Unix) to a throwaway temp dir so token/credential tests never touch the real
// user config.
func isolateConfigDir(t *testing.T) {
	t.Helper()
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
}

func TestSaveLoadTokenRoundtrip(t *testing.T) {
	isolateConfigDir(t)
	t.Setenv("SSHM_GOOGLE_CLIENT_ID", "")
	t.Setenv("SSHM_GOOGLE_CLIENT_SECRET", "")

	want := &oauth2.Token{
		AccessToken:  "access-123",
		TokenType:    "Bearer",
		RefreshToken: "refresh-456",
		Expiry:       time.Now().Add(time.Hour).Truncate(time.Second),
	}
	if err := SaveToken(want); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}
	if !HasToken() {
		t.Fatal("HasToken = false after SaveToken")
	}

	got, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken: %v", err)
	}
	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken {
		t.Errorf("token roundtrip mismatch: %+v vs %+v", got, want)
	}
	if !got.Expiry.Equal(want.Expiry) {
		t.Errorf("expiry = %v, want %v", got.Expiry, want.Expiry)
	}
}

func TestLoadTokenMissing(t *testing.T) {
	isolateConfigDir(t)
	if _, err := LoadToken(); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("LoadToken error = %v, want os.ErrNotExist", err)
	}
	if HasToken() {
		t.Error("HasToken = true with no saved token")
	}
}

func TestLoadCredentialsFromEnv(t *testing.T) {
	t.Setenv("SSHM_GOOGLE_CLIENT_ID", "client-id")
	t.Setenv("SSHM_GOOGLE_CLIENT_SECRET", "client-secret")

	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds.ClientID != "client-id" || creds.ClientSecret != "client-secret" {
		t.Errorf("credentials = %+v, want the env values", creds)
	}
}

func TestLoadCredentialsMissing(t *testing.T) {
	isolateConfigDir(t)
	t.Setenv("SSHM_GOOGLE_CLIENT_ID", "")
	t.Setenv("SSHM_GOOGLE_CLIENT_SECRET", "")

	if _, err := LoadCredentials(); err == nil {
		t.Fatal("LoadCredentials error = nil, want a missing-credentials error")
	}
}

func TestSaveLoadCredentialsRoundtrip(t *testing.T) {
	isolateConfigDir(t)
	t.Setenv("SSHM_GOOGLE_CLIENT_ID", "")
	t.Setenv("SSHM_GOOGLE_CLIENT_SECRET", "")

	want := Credentials{ClientID: "cid.apps.googleusercontent.com", ClientSecret: "shh"}
	if err := SaveCredentials(want); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	got, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if got != want {
		t.Errorf("credentials = %+v, want %+v", got, want)
	}
}

func TestSaveCredentialsFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix file permissions not meaningful on Windows")
	}
	isolateConfigDir(t)
	t.Setenv("SSHM_GOOGLE_CLIENT_ID", "")
	t.Setenv("SSHM_GOOGLE_CLIENT_SECRET", "")

	if err := SaveCredentials(Credentials{ClientID: "cid", ClientSecret: "s"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
	path, err := CredentialsPath()
	if err != nil {
		t.Fatalf("CredentialsPath: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("credentials file mode = %o, want 0600", perm)
	}
}
