package drive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Gu1llaum-3/sshm/internal/config"
	"golang.org/x/oauth2"
)

// AppDataScope grants access only to this app's hidden appDataFolder, not to
// the user's regular Drive files.
const AppDataScope = "https://www.googleapis.com/auth/drive.appdata"

// googleEndpoint is defined locally to avoid pulling golang.org/x/oauth2/google
// (and its cloud metadata dependency) just for two URLs.
var googleEndpoint = oauth2.Endpoint{
	AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
	TokenURL:  "https://oauth2.googleapis.com/token",
	AuthStyle: oauth2.AuthStyleInParams,
}

// Credentials are the OAuth client id/secret of a Google Cloud "Desktop app".
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// storedToken is the on-disk representation of an oauth2.Token.
type storedToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// TokenPath returns the path of the saved OAuth token.
func TokenPath() (string, error) {
	dir, err := config.GetSSHMConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "drive_token.json"), nil
}

// CredentialsPath returns the path of the optional client credentials file.
func CredentialsPath() (string, error) {
	dir, err := config.GetSSHMConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "drive_client.json"), nil
}

// LoadCredentials returns the OAuth client credentials, in order of
// precedence: the SSHM_GOOGLE_CLIENT_ID/SECRET env vars, the per-user
// drive_client.json file, or the credentials baked into official builds.
func LoadCredentials() (Credentials, error) {
	if id := os.Getenv("SSHM_GOOGLE_CLIENT_ID"); id != "" {
		return Credentials{
			ClientID:     id,
			ClientSecret: os.Getenv("SSHM_GOOGLE_CLIENT_SECRET"),
		}, nil
	}

	path, err := CredentialsPath()
	if err != nil {
		return Credentials{}, err
	}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		var creds Credentials
		if uerr := json.Unmarshal(data, &creds); uerr != nil {
			return Credentials{}, uerr
		}
		if creds.ClientID == "" {
			return Credentials{}, fmt.Errorf("credentials file %s has no client_id", path)
		}
		return creds, nil
	case !errors.Is(err, os.ErrNotExist):
		return Credentials{}, err
	}

	if builtinClientID != "" {
		return Credentials{ClientID: builtinClientID, ClientSecret: builtinClientSecret}, nil
	}
	return Credentials{}, fmt.Errorf("no Google credentials: set SSHM_GOOGLE_CLIENT_ID/SECRET, create %s, or use an official sshm build", path)
}

// SaveCredentials writes the OAuth client credentials to the per-user
// credentials file with 0600 permissions.
func SaveCredentials(creds Credentials) error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SaveToken writes the token to disk with 0600 permissions.
func SaveToken(tok *oauth2.Token) error {
	path, err := TokenPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(storedToken{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadToken reads the saved token, or returns os.ErrNotExist when absent.
func LoadToken() (*oauth2.Token, error) {
	path, err := TokenPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var st storedToken
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken:  st.AccessToken,
		TokenType:    st.TokenType,
		RefreshToken: st.RefreshToken,
		Expiry:       st.Expiry,
	}, nil
}

// HasToken reports whether a token has been saved.
func HasToken() bool {
	_, err := LoadToken()
	return err == nil
}

func oauthConfig(creds Credentials, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Scopes:       []string{AppDataScope},
		Endpoint:     googleEndpoint,
		RedirectURL:  redirectURL,
	}
}

// AuthenticatedClient returns an HTTP client that refreshes the saved token as
// needed.
func AuthenticatedClient(ctx context.Context) (*http.Client, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	tok, err := LoadToken()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("not logged in: run 'sshm drive login' first")
		}
		return nil, err
	}
	return oauth2.NewClient(ctx, oauthConfig(creds, "").TokenSource(ctx, tok)), nil
}

// LoginFlow runs the loopback OAuth flow with the given client credentials,
// saving the resulting token. The authorization URL is written to out and the
// browser is opened best-effort.
func LoginFlow(ctx context.Context, creds Credentials, out io.Writer) error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer ln.Close()

	redirect := fmt.Sprintf("http://127.0.0.1:%d/callback", ln.Addr().(*net.TCPAddr).Port)
	cfg := oauthConfig(creds, redirect)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if e := r.URL.Query().Get("error"); e != "" {
			fmt.Fprintf(w, "Authorization failed: %s", e)
			errCh <- fmt.Errorf("authorization denied: %s", e)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, "Login complete. You can close this tab and return to the terminal.")
		codeCh <- code
	})}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	authURL := cfg.AuthCodeURL("sshm-state",
		oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Fprintf(out, "Open this URL in your browser to authorize sshm:\n\n  %s\n\n", authURL)
	openBrowser(authURL)

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Minute):
		return errors.New("timed out waiting for authorization")
	case <-ctx.Done():
		return ctx.Err()
	}

	tok, err := cfg.Exchange(context.WithValue(ctx, oauth2.HTTPClient, http.DefaultClient), code)
	if err != nil {
		return fmt.Errorf("token exchange: %w", err)
	}
	if err := SaveToken(tok); err != nil {
		return err
	}
	path, _ := TokenPath()
	fmt.Fprintf(out, "Logged in. Token saved to %s\n", path)
	return nil
}

// openBrowser tries to open url with the OS default browser (best effort).
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
