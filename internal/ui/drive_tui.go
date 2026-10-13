package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nicozilabe/sshm/internal/config"
	"github.com/Nicozilabe/sshm/internal/drive"
	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) startDriveSyncCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		remote, err := newDriveRemoteForTUI(ctx)
		if err != nil {
			return driveSyncMsg{err: err}
		}

		path, err := resolveSSHConfigPathForTUI(m.configFile)
		if err != nil {
			return driveSyncMsg{err: err}
		}

		local, err := readFileOrEmptyForTUI(path)
		if err != nil {
			return driveSyncMsg{err: err}
		}

		res, err := drive.Sync(remote, local)
		if err != nil {
			return driveSyncMsg{err: err}
		}

	_, err = writeWithBackupTUI(path, res.Content)
	if err != nil {
		return driveSyncMsg{err: err}
	}

		return driveSyncMsg{result: &res}
	}
}

func (m Model) startDriveLoginCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		cmd := &cobra.Command{}
		creds, err := drive.LoadCredentials()
		if err != nil {
			creds, err = promptCredentialsForTUI(cmd)
			if err != nil {
				return driveLoginMsg{err: err}
			}
			if err := drive.SaveCredentials(creds); err != nil {
			}
		}
		if err := drive.LoginFlow(ctx, creds, io.Discard); err != nil {
			return driveLoginMsg{err: err}
		}
		return driveLoginMsg{}
	}
}
func newDriveRemoteForTUI(ctx context.Context) (*drive.Drive, error) {
	client, err := drive.AuthenticatedClient(ctx)
	if err != nil {
		return nil, err
	}
	return drive.NewDrive(client), nil
}

func resolveSSHConfigPathForTUI(configFile string) (string, error) {
	if configFile != "" {
		return configFile, nil
	}
	return config.GetDefaultSSHConfigPath()
}

func readFileOrEmptyForTUI(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}



func promptCredentialsForTUI(cmd *cobra.Command) (drive.Credentials, error) {
	creds := drive.Credentials{}
	fmt.Fprint(os.Stdout, "Client ID: ")
	id, err := readLineForTUI(os.Stdin)
	if err != nil {
		return creds, err
	}
	fmt.Fprint(os.Stdout, "Client Secret: ")
	secret, err := readSecretForTUI(os.Stdin)
	if err != nil {
		return creds, err
	}
	fmt.Fprintln(os.Stdout)
	creds.ClientID = strings.TrimSpace(id)
	creds.ClientSecret = strings.TrimSpace(secret)
	if creds.ClientID == "" {
		return drive.Credentials{}, fmt.Errorf("client id is required")
	}
	return creds, nil
}

func readLineForTUI(r io.Reader) (string, error) {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if buf[0] == 10 {
				return b.String(), nil
			}
			if buf[0] == 13 {
				continue
			}
			b.WriteByte(buf[0])
		}
		if err != nil {
			if err == io.EOF {
				return b.String(), nil
			}
			return b.String(), err
		}
	}
}

func readSecretForTUI(r io.Reader) (string, error) {
	return readLineForTUI(r)
}

func writeWithBackupTUI(path string, data []byte) (string, error) {
	backup := ""
	if existing, err := os.ReadFile(path); err == nil {
		if dir, derr := config.GetSSHMBackupDir(); derr == nil {
			if merr := os.MkdirAll(dir, 0700); merr == nil {
				backup = filepath.Join(dir, fmt.Sprintf("config-%d.bak", time.Now().Unix()))
				if werr := os.WriteFile(backup, existing, 0600); werr != nil {
					backup = ""
				}
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return backup, err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return backup, err
	}
	return backup, nil
}
