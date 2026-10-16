// Package updater implements the self-update logic behind "baq update": it
// resolves a release, downloads and verifies the matching archive and swaps the
// running binary in place.
package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Nicozilabe/baqueano/internal/version"
)

// defaultHTTPTimeout bounds a whole download; release archives are a few MB.
const defaultHTTPTimeout = 5 * time.Minute

// Options configures a single update run.
type Options struct {
	// CurrentVersion is the running build's version (cmd.AppVersion).
	CurrentVersion string
	// TargetVersion, when set, pins the exact release tag (e.g. "v1.2.3").
	// An empty value means the latest stable release.
	TargetVersion string
	// IncludePrerelease allows a pre-release as the target when no explicit
	// version is given.
	IncludePrerelease bool
	// Force reinstalls even when the current version is already up to date.
	Force bool
	// CheckOnly only reports whether an update is available.
	CheckOnly bool
	// Out receives progress messages; defaults to io.Discard.
	Out io.Writer
}

// Result describes the outcome of an update run.
type Result struct {
	From       string
	To         string
	Checked    bool
	Updated    bool
	AssetName  string
	Path       string
	ReleaseURL string
}

// Run performs the update described by opts.
func Run(ctx context.Context, opts Options) (*Result, error) {
	out := opts.Out
	if out == nil {
		out = io.Discard
	}
	if opts.CurrentVersion == "" {
		opts.CurrentVersion = "dev"
	}

	client := &http.Client{Timeout: defaultHTTPTimeout}

	release, err := resolveRelease(ctx, opts)
	if err != nil {
		return nil, err
	}

	res := &Result{
		From:       opts.CurrentVersion,
		To:         release.TagName,
		ReleaseURL: release.HTMLURL,
	}

	upToDate := opts.CurrentVersion != "dev" &&
		version.Compare(opts.CurrentVersion, release.TagName) >= 0

	if upToDate && !opts.Force {
		res.Checked = true
		fmt.Fprintf(out, "baq is already up to date (%s).\n", opts.CurrentVersion)
		return res, nil
	}

	if opts.CheckOnly {
		res.Checked = true
		fmt.Fprintf(out, "Update available: %s (current %s).\n", release.TagName, opts.CurrentVersion)
		return res, nil
	}

	assetName := AssetName(runtime.GOOS, runtime.GOARCH)
	res.AssetName = assetName

	asset, err := findAsset(release, assetName)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(out, "Downloading %s...\n", assetName)
	archive, err := download(ctx, client, asset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", assetName, err)
	}

	if err := verifyChecksum(ctx, client, release, assetName, archive); err != nil {
		return nil, err
	}

	binary, err := extractBinary(assetName, archive)
	if err != nil {
		return nil, err
	}

	path, err := replaceExecutable(binary)
	if err != nil {
		return nil, err
	}

	res.Updated = true
	res.Path = path
	fmt.Fprintf(out, "Updated baq from %s to %s (%s).\n", opts.CurrentVersion, release.TagName, path)
	fmt.Fprintln(out, "Restart baq to use the new version.")
	return res, nil
}

// resolveRelease picks the release to target based on the options.
func resolveRelease(ctx context.Context, opts Options) (*version.GitHubRelease, error) {
	ua := "baq/" + opts.CurrentVersion

	if opts.TargetVersion != "" {
		return version.FetchRelease(ctx, opts.TargetVersion, ua)
	}

	if opts.IncludePrerelease {
		releases, err := version.FetchReleases(ctx, ua)
		if err != nil {
			return nil, err
		}
		for i := range releases {
			if !releases[i].Draft {
				return &releases[i], nil
			}
		}
		return nil, errors.New("no releases found")
	}

	return version.FetchRelease(ctx, "", ua)
}

// findAsset returns the asset matching name (case-insensitive).
func findAsset(release *version.GitHubRelease, name string) (version.GitHubReleaseAsset, error) {
	for _, a := range release.Assets {
		if strings.EqualFold(a.Name, name) {
			return a, nil
		}
	}
	var available []string
	for _, a := range release.Assets {
		available = append(available, a.Name)
	}
	return version.GitHubReleaseAsset{}, fmt.Errorf("release %s has no asset %q (available: %s)",
		release.TagName, name, strings.Join(available, ", "))
}

// download GETs a URL and returns the whole body.
func download(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "baq-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// verifyChecksum checks archive against the release's checksums.txt.
func verifyChecksum(ctx context.Context, client *http.Client, release *version.GitHubRelease, assetName string, archive []byte) error {
	asset, err := findAsset(release, "checksums.txt")
	if err != nil {
		return err
	}
	sums, err := download(ctx, client, asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("download checksums.txt: %w", err)
	}

	expected, err := checksumFor(sums, assetName)
	if err != nil {
		return err
	}

	actual := sha256.Sum256(archive)
	got := hex.EncodeToString(actual[:])
	if !strings.EqualFold(expected, got) {
		return fmt.Errorf("checksum mismatch for %s (expected %s, got %s)", assetName, expected, got)
	}
	return nil
}

// checksumFor extracts the "<sha256>  <filename>" entry for filename.
func checksumFor(checksums []byte, filename string) (string, error) {
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("no checksum found for %s", filename)
}

// extractBinary pulls the baq executable out of a release archive.
func extractBinary(assetName string, archive []byte) ([]byte, error) {
	if strings.HasSuffix(strings.ToLower(assetName), ".zip") {
		return extractFromZip(archive)
	}
	return extractFromTarGz(archive)
}

func extractFromTarGz(archive []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && isBinaryName(hdr.Name) {
			return io.ReadAll(tr)
		}
	}
	return nil, errors.New("baq executable not found in archive")
}

func extractFromZip(archive []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !isBinaryName(f.Name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, errors.New("baq executable not found in archive")
}

// isBinaryName reports whether an archive entry is the baq executable.
func isBinaryName(name string) bool {
	base := filepath.Base(name)
	return base == "baq" || base == "baq.exe"
}

// AssetName mirrors the GoReleaser archive name template, e.g.
// "baqueano_Linux_x86_64.tar.gz" or "baqueano_Windows_x86_64.zip".
func AssetName(goos, goarch string) string {
	osName := map[string]string{
		"darwin":  "Darwin",
		"linux":   "Linux",
		"windows": "Windows",
	}[goos]
	if osName == "" {
		osName = goos
	}

	archName := map[string]string{
		"amd64": "x86_64",
		"386":   "i386",
		"arm":   "armv7",
	}[goarch]
	if archName == "" {
		archName = goarch
	}

	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("baqueano_%s_%s.%s", osName, archName, ext)
}

// replaceExecutable atomically installs binary over the running executable and
// returns the resulting path. On Windows the running file cannot be overwritten
// in place, so it is renamed aside first and removed on the next launch.
func replaceExecutable(binary []byte) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate current executable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	dir := filepath.Dir(exe)

	tmp, err := os.CreateTemp(dir, ".baq-new-*")
	if err != nil {
		return "", fmt.Errorf("cannot write to %s (try the install script or a directory you own): %w", dir, err)
	}
	tmpName := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(binary); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpName, 0o755); err != nil {
			return "", err
		}
	}

	if runtime.GOOS == "windows" {
		old := exe + ".old"
		_ = os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			return "", fmt.Errorf("failed to move current binary aside: %w", err)
		}
		if err := os.Rename(tmpName, exe); err != nil {
			_ = os.Rename(old, exe)
			return "", fmt.Errorf("failed to install new binary: %w", err)
		}
		removeTmp = false
		// Best effort: this only succeeds once the old process has exited.
		_ = os.Remove(old)
		return exe, nil
	}

	if err := os.Rename(tmpName, exe); err != nil {
		return "", fmt.Errorf("failed to install new binary: %w", err)
	}
	removeTmp = false
	return exe, nil
}

// CleanupOldBinary removes leftovers from a previous Windows self-update. It is
// safe to call unconditionally and ignores errors.
func CleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	_ = os.Remove(exe + ".old")

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(exe), ".baq-new-*"))
	if err != nil {
		return
	}
	for _, m := range matches {
		_ = os.Remove(m)
	}
}
