package config

import (
	"os/exec"
	"strings"
	"testing"
)

// sshG runs `ssh -F config -G host` and returns the output with the CRLF line
// endings that OpenSSH emits on Windows normalized away, so token checks are
// cross-platform.
func sshG(t *testing.T, configFile, host string) []byte {
	t.Helper()
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		t.Skip("ssh client not installed")
	}
	out, err := exec.Command(sshPath, "-F", configFile, "-G", host).CombinedOutput()
	if err != nil {
		t.Fatalf("ssh -G failed: %v\n%s", err, out)
	}
	return []byte(strings.ReplaceAll(string(out), "\r", ""))
}