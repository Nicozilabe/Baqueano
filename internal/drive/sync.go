package drive

import "fmt"

// Remote stores a single config document. Implementations only need to move
// bytes; all merging happens in the pure layer so it can be tested offline.
type Remote interface {
	// Download returns the stored config, or (nil, nil) when nothing has been
	// stored yet.
	Download() ([]byte, error)
	// Upload overwrites the stored config.
	Upload(data []byte) error
}

// Result is the outcome of a sync operation.
type Result struct {
	// Content is the merged document that was produced (and uploaded).
	Content []byte
	// Stats describes how many hosts were merged.
	Stats Stats
	// RemoteEmpty is true when the remote had no document before this sync.
	RemoteEmpty bool
}

// Sync downloads the remote document, merges it with the local one (local
// wins), uploads the result and returns it.
func Sync(r Remote, local []byte) (Result, error) {
	remote, err := r.Download()
	if err != nil {
		return Result{}, fmt.Errorf("download from remote: %w", err)
	}

	remoteEmpty := len(trimSpace(remote)) == 0
	merged, stats := Merge(local, remote)

	if err := r.Upload(merged); err != nil {
		return Result{}, fmt.Errorf("upload to remote: %w", err)
	}

	return Result{Content: merged, Stats: stats, RemoteEmpty: remoteEmpty}, nil
}

// Pull returns the remote document without merging.
func Pull(r Remote) ([]byte, error) {
	remote, err := r.Download()
	if err != nil {
		return nil, fmt.Errorf("download from remote: %w", err)
	}
	return remote, nil
}

// Push overwrites the remote document with local, without merging.
func Push(r Remote, local []byte) error {
	if err := r.Upload(local); err != nil {
		return fmt.Errorf("upload to remote: %w", err)
	}
	return nil
}

// trimSpace is a tiny helper so callers don't need to import strings just to
// detect an empty remote document (including whitespace-only files).
func trimSpace(b []byte) []byte {
	start := 0
	for start < len(b) && isSpace(b[start]) {
		start++
	}
	end := len(b)
	for end > start && isSpace(b[end-1]) {
		end--
	}
	return b[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
