// Package drive implements an experimental sync backend that stores the SSH
// host list in the user's Google Drive (hidden appDataFolder) so the same
// config can be shared across machines.
//
// It is deliberately split in two layers: a pure merge engine (this file and
// sync.go) that is fully testable without any network, and a thin Google Drive
// client (drive.go/oauth.go) that only moves bytes around.
package drive

import (
	"strings"
)

// Stats summarizes what a merge did.
type Stats struct {
	// Local is the number of host blocks present locally.
	Local int
	// Remote is the number of host blocks present remotely.
	Remote int
	// Added is the number of remote-only host blocks appended to the result.
	Added int
	// Conflict is the number of host blocks present on both sides (local wins).
	Conflict int
}

// block is a chunk of ssh config text. A block with an empty key is the
// preamble (global directives, comments) that appears before the first
// Host/Match. Every other block starts with a Host line; Match blocks get a
// synthetic key so they never collide with hosts.
type block struct {
	key   string
	lines []string
}

// isBlockStart reports whether a line opens a new host/match block.
func isBlockStart(line string) (kind, key string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", ""
	}
	switch strings.ToLower(fields[0]) {
	case "host":
		names := make([]string, 0, len(fields)-1)
		for _, f := range fields[1:] {
			names = append(names, strings.ToLower(f))
		}
		return "host", strings.Join(names, ",")
	case "match":
		// Match blocks are kept from the local side only; give them a unique
		// key derived from the full directive so remote matches can't merge in.
		return "match", "match " + strings.ToLower(strings.TrimSpace(line[len(fields[0]):]))
	default:
		return "", ""
	}
}

// parseBlocks splits config text into a preamble block (key "") plus one block
// per Host or Match directive. CRLF is normalized to LF.
func parseBlocks(text string) []block {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")

	var blocks []block
	cur := block{key: ""}
	for _, line := range lines {
		if kind, key := isBlockStart(line); kind != "" {
			blocks = append(blocks, cur)
			cur = block{key: key}
			_ = kind
		}
		cur.lines = append(cur.lines, line)
	}
	blocks = append(blocks, cur)
	return blocks
}

// renderBlocks joins blocks back into text, trimming the trailing blank lines
// that the split leaves behind and ensuring a single final newline.
func renderBlocks(blocks []block) string {
	var out []string
	for _, b := range blocks {
		out = append(out, b.lines...)
	}
	// Drop trailing blank lines.
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n") + "\n"
}

// Merge unions the host blocks of local and remote text. Local always wins on
// conflicts, the local preamble and Match blocks are preserved, and remote-only
// host blocks are appended at the end. The result is idempotent: merging a
// merged file with the same remote yields the same bytes.
func Merge(local, remote []byte) ([]byte, Stats) {
	localBlocks := parseBlocks(string(local))
	remoteBlocks := parseBlocks(string(remote))

	var stats Stats
	seen := make(map[string]bool)
	result := make([]block, 0, len(localBlocks))

	for _, b := range localBlocks {
		if b.key != "" && !strings.HasPrefix(b.key, "match ") {
			stats.Local++
			seen[b.key] = true
		}
		result = append(result, b)
	}

	for _, b := range remoteBlocks {
		if b.key == "" || strings.HasPrefix(b.key, "match ") {
			continue
		}
		stats.Remote++
		if seen[b.key] {
			stats.Conflict++
			continue
		}
		seen[b.key] = true
		stats.Added++
		result = append(result, b)
	}

	return []byte(renderBlocks(result)), stats
}
