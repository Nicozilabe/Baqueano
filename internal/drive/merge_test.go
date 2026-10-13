package drive

import "testing"

const (
	localConfig = `Host alpha
  HostName alpha.example.com
  User alice

Host beta
  HostName beta.example.com
`
	remoteConfig = `Host beta
  HostName beta-remote.example.com
  User bob

Host gamma
  HostName gamma.example.com
`
)

func TestMergeUnionsHosts(t *testing.T) {
	merged, stats := Merge([]byte(localConfig), []byte(remoteConfig))

	if stats.Local != 2 {
		t.Errorf("Local = %d, want 2", stats.Local)
	}
	if stats.Remote != 2 {
		t.Errorf("Remote = %d, want 2", stats.Remote)
	}
	if stats.Added != 1 {
		t.Errorf("Added = %d, want 1", stats.Added)
	}
	if stats.Conflict != 1 {
		t.Errorf("Conflict = %d, want 1", stats.Conflict)
	}

	got := string(merged)
	for _, want := range []string{"Host alpha", "Host beta", "Host gamma"} {
		if !containsLine(got, want) {
			t.Errorf("merged config missing %q\n%s", want, got)
		}
	}
}

func TestMergeLocalWinsOnConflict(t *testing.T) {
	merged, _ := Merge([]byte(localConfig), []byte(remoteConfig))
	got := string(merged)

	if containsLine(got, "HostName beta-remote.example.com") {
		t.Errorf("remote beta content leaked into merged config:\n%s", got)
	}
	if !containsLine(got, "HostName beta.example.com") {
		t.Errorf("local beta content missing:\n%s", got)
	}
}

func TestMergeIsIdempotent(t *testing.T) {
	first, _ := Merge([]byte(localConfig), []byte(remoteConfig))
	second, stats := Merge(first, []byte(remoteConfig))

	if string(first) != string(second) {
		t.Errorf("merge is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if stats.Added != 0 || stats.Conflict != 2 {
		t.Errorf("second merge Added=%d Conflict=%d, want 0/2", stats.Added, stats.Conflict)
	}
}

func TestMergeEmptyRemoteReturnsLocal(t *testing.T) {
	merged, stats := Merge([]byte(localConfig), nil)
	if string(merged) != localConfig {
		t.Errorf("merge with empty remote changed the config:\n%s", merged)
	}
	if stats.Added != 0 {
		t.Errorf("Added = %d, want 0", stats.Added)
	}
}

func TestMergePreservesPreambleAndMatch(t *testing.T) {
	local := `# global options
ServerAliveInterval 60

Match host *.internal
  User svc

Host alpha
  HostName alpha.example.com
`
	remote := `Host gamma
  HostName gamma.example.com
`
	merged, _ := Merge([]byte(local), []byte(remote))
	got := string(merged)

	for _, want := range []string{
		"# global options",
		"ServerAliveInterval 60",
		"Match host *.internal",
		"User svc",
		"Host alpha",
		"Host gamma",
	} {
		if !containsLine(got, want) {
			t.Errorf("merged config missing %q\n%s", want, got)
		}
	}
}

func TestMergeNormalizesCRLF(t *testing.T) {
	local := "Host alpha\r\n  HostName alpha.example.com\r\n"
	remote := "Host beta\r\n  HostName beta.example.com\r\n"
	merged, _ := Merge([]byte(local), []byte(remote))
	got := string(merged)

	if contains(got, "\r") {
		t.Errorf("merged config still contains CR characters:\n%q", got)
	}
	if !containsLine(got, "Host alpha") || !containsLine(got, "Host beta") {
		t.Errorf("merged config lost hosts:\n%s", got)
	}
}

func TestMergeRemoteMatchIgnored(t *testing.T) {
	local := "Host alpha\n  HostName alpha.example.com\n"
	remote := "Match all\n  ForwardAgent yes\n"
	merged, _ := Merge([]byte(local), []byte(remote))
	if containsLine(string(merged), "ForwardAgent") {
		t.Errorf("remote Match block should not be merged:\n%s", merged)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// containsLine reports whether any line of s, trimmed, equals want.
func containsLine(s, want string) bool {
	for _, line := range splitLines(s) {
		if trimSpaceStr(line) == want {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func trimSpaceStr(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
