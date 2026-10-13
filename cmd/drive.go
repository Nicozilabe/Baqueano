package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nicozilabe/sshm/internal/config"
	"github.com/Nicozilabe/sshm/internal/drive"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// driveCmd groups the experimental Google Drive sync commands.
var driveCmd = &cobra.Command{
	Use:   "drive",
	Short: "Sync your SSH hosts through Google Drive (experimental)",
	Long: `Sync your SSH config through your Google Drive account (experimental).

The hosts are stored in a hidden appDataFolder that only this app can see, so
they never mix with your regular Drive files. Only the SSH config document is
synced: private keys are never uploaded.

First create a Google Cloud "Desktop app" OAuth client. Then just run login: if
SSHM_GOOGLE_CLIENT_ID / SSHM_GOOGLE_CLIENT_SECRET are not set and no
<config-dir>/drive_client.json exists, sshm asks for the client id and secret
and saves them for next time.

  sshm drive login     # one-time browser authorization
  sshm drive sync      # merge local + remote (local wins) and upload`,
}

var driveSyncDryRun bool

var (
	driveLoginClientID     string
	driveLoginClientSecret string
)

var driveLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorize sshm to access its Google Drive app folder",
	Long: `Authorize sshm to access its Google Drive app folder.

The OAuth client id and secret are taken from SSHM_GOOGLE_CLIENT_ID /
SSHM_GOOGLE_CLIENT_SECRET, from <config-dir>/drive_client.json, or from the
--client-id/--client-secret flags. If none are available, you are prompted and
the values are saved to <config-dir>/drive_client.json.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := resolveLoginCredentials(cmd)
		if err != nil {
			return err
		}
		return drive.LoginFlow(cmd.Context(), creds, cmd.OutOrStdout())
	},
}

var driveSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Merge local and remote hosts, then upload the result",
	Long: `Download the remote config, merge it with your local one (local hosts win
on conflicts, remote-only hosts are appended), write the merged file locally
and upload it back.

A backup of the current local config is written to the sshm backup directory
before it is overwritten.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		remote, err := newDriveRemote(ctx)
		if err != nil {
			return err
		}

		path, err := resolveSSHConfigPath()
		if err != nil {
			return err
		}
		local, err := readFileOrEmpty(path)
		if err != nil {
			return err
		}

		res, err := drive.Sync(remote, local)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Merged: %d local, %d remote - %d added, %d conflicts (local kept)\n",
			res.Stats.Local, res.Stats.Remote, res.Stats.Added, res.Stats.Conflict)

		if driveSyncDryRun {
			fmt.Fprintln(out, "Dry run: no files written.")
			return nil
		}

		backup, err := writeWithBackup(path, res.Content)
		if err != nil {
			return err
		}
		if backup != "" {
			fmt.Fprintf(out, "Backup saved to %s\n", backup)
		}
		fmt.Fprintf(out, "Updated %s\n", path)
		return nil
	},
}

var drivePullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Overwrite the local config with the remote one (no merge)",
	RunE: func(cmd *cobra.Command, args []string) error {
		remote, err := newDriveRemote(cmd.Context())
		if err != nil {
			return err
		}
		data, err := drive.Pull(remote)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("remote config is empty; nothing to pull")
		}
		path, err := resolveSSHConfigPath()
		if err != nil {
			return err
		}
		backup, err := writeWithBackup(path, data)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if backup != "" {
			fmt.Fprintf(out, "Backup saved to %s\n", backup)
		}
		fmt.Fprintf(out, "Updated %s from remote\n", path)
		return nil
	},
}

var drivePushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload the local config, overwriting the remote one (no merge)",
	RunE: func(cmd *cobra.Command, args []string) error {
		remote, err := newDriveRemote(cmd.Context())
		if err != nil {
			return err
		}
		path, err := resolveSSHConfigPath()
		if err != nil {
			return err
		}
		local, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := drive.Push(remote, local); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Uploaded %s to Google Drive\n", path)
		return nil
	},
}

// newDriveRemote builds an authenticated Drive remote.
func newDriveRemote(ctx context.Context) (*drive.Drive, error) {
	client, err := drive.AuthenticatedClient(ctx)
	if err != nil {
		return nil, err
	}
	return drive.NewDrive(client), nil
}

// resolveLoginCredentials returns the OAuth client credentials for the login
// flow, in order of precedence: flags, environment/file, interactive prompt.
// Credentials gathered from flags or the prompt are saved for later commands.
func resolveLoginCredentials(cmd *cobra.Command) (drive.Credentials, error) {
	if driveLoginClientID != "" {
		return saveCredentials(cmd, drive.Credentials{
			ClientID:     driveLoginClientID,
			ClientSecret: driveLoginClientSecret,
		})
	}

	if creds, err := drive.LoadCredentials(); err == nil {
		return creds, nil
	}

	creds, err := promptCredentials(cmd)
	if err != nil {
		return drive.Credentials{}, err
	}
	return saveCredentials(cmd, creds)
}

// saveCredentials persists credentials to disk (best effort) and reports where,
// so that later non-interactive commands like sync can reuse them.
func saveCredentials(cmd *cobra.Command, creds drive.Credentials) (drive.Credentials, error) {
	creds.ClientID = strings.TrimSpace(creds.ClientID)
	creds.ClientSecret = strings.TrimSpace(creds.ClientSecret)
	if creds.ClientID == "" {
		return drive.Credentials{}, errors.New("client id is required")
	}
	if err := drive.SaveCredentials(creds); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not save credentials: %v\n", err)
		return creds, nil
	}
	if path, err := drive.CredentialsPath(); err == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Saved credentials to %s\n", path)
	}
	return creds, nil
}

// promptCredentials asks the user for the OAuth client id and secret.
func promptCredentials(cmd *cobra.Command) (drive.Credentials, error) {
	in := cmd.InOrStdin()
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "No Google credentials found.")
	fmt.Fprintln(out, "Create an OAuth client of type \"Desktop app\" at")
	fmt.Fprintln(out, "  https://console.cloud.google.com/apis/credentials")
	fmt.Fprintln(out)

	fmt.Fprint(out, "Client ID: ")
	id, err := readLine(in)
	if err != nil {
		return drive.Credentials{}, err
	}

	fmt.Fprint(out, "Client Secret: ")
	secret, err := readSecret(in)
	if err != nil {
		return drive.Credentials{}, err
	}
	fmt.Fprintln(out)

	return drive.Credentials{
		ClientID:     strings.TrimSpace(id),
		ClientSecret: strings.TrimSpace(secret),
	}, nil
}

// readLine reads a single line, stripping the trailing newline and any CR. It
// reads byte by byte so it never consumes input meant for the next prompt.
func readLine(r io.Reader) (string, error) {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			switch buf[0] {
			case '\n':
				return b.String(), nil
			case '\r':
				// ignore
			default:
				b.WriteByte(buf[0])
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return b.String(), nil
			}
			return b.String(), err
		}
	}
}

// readSecret reads a secret line, hiding it from the terminal when input comes
// from a real terminal and falling back to a visible line otherwise (tests,
// pipes).
func readSecret(r io.Reader) (string, error) {
	if f, ok := r.(*os.File); ok {
		fd := int(f.Fd())
		if term.IsTerminal(fd) {
			data, err := term.ReadPassword(fd)
			return string(data), err
		}
	}
	return readLine(r)
}

// resolveSSHConfigPath returns the -c flag value or the default ~/.ssh/config.
func resolveSSHConfigPath() (string, error) {
	if configFile != "" {
		return configFile, nil
	}
	return config.GetDefaultSSHConfigPath()
}

func readFileOrEmpty(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

// writeWithBackup snapshots an existing file into the sshm backup dir before
// overwriting it, and returns the backup path ("" when there was nothing to
// back up).
func writeWithBackup(path string, data []byte) (string, error) {
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

func init() {
	driveSyncCmd.Flags().BoolVar(&driveSyncDryRun, "dry-run", false, "Show what would change without writing or uploading")
	driveLoginCmd.Flags().StringVar(&driveLoginClientID, "client-id", "", "OAuth client id (skips the interactive prompt)")
	driveLoginCmd.Flags().StringVar(&driveLoginClientSecret, "client-secret", "", "OAuth client secret (use with --client-id)")
	driveCmd.AddCommand(driveLoginCmd, driveSyncCmd, drivePullCmd, drivePushCmd)
	RootCmd.AddCommand(driveCmd)
}
