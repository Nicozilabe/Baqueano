package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/Nicozilabe/baqueano/internal/version"
)

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos, goarch, want string
	}{
		{"linux", "amd64", "baqueano_Linux_x86_64.tar.gz"},
		{"linux", "arm64", "baqueano_Linux_arm64.tar.gz"},
		{"linux", "386", "baqueano_Linux_i386.tar.gz"},
		{"linux", "arm", "baqueano_Linux_armv7.tar.gz"},
		{"darwin", "amd64", "baqueano_Darwin_x86_64.tar.gz"},
		{"darwin", "arm64", "baqueano_Darwin_arm64.tar.gz"},
		{"windows", "amd64", "baqueano_Windows_x86_64.zip"},
		{"windows", "386", "baqueano_Windows_i386.zip"},
	}
	for _, tt := range tests {
		if got := AssetName(tt.goos, tt.goarch); got != tt.want {
			t.Errorf("AssetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestChecksumFor(t *testing.T) {
	sums := []byte("aaa  baqueano_Linux_x86_64.tar.gz\nbbb  checksums.txt\n")

	got, err := checksumFor(sums, "baqueano_Linux_x86_64.tar.gz")
	if err != nil {
		t.Fatalf("checksumFor returned error: %v", err)
	}
	if got != "aaa" {
		t.Errorf("checksumFor = %q, want %q", got, "aaa")
	}

	if _, err := checksumFor(sums, "missing.tar.gz"); err == nil {
		t.Error("expected error for missing checksum, got nil")
	}
}

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0o755,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("write body: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func makeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func TestExtractFromTarGz(t *testing.T) {
	archive := makeTarGz(t, map[string]string{
		"LICENSE":   "license",
		"baq":       "BINARY-CONTENT",
		"README.md": "readme",
	})

	got, err := extractFromTarGz(archive)
	if err != nil {
		t.Fatalf("extractFromTarGz: %v", err)
	}
	if string(got) != "BINARY-CONTENT" {
		t.Errorf("extracted %q, want %q", got, "BINARY-CONTENT")
	}
}

func TestExtractFromZip(t *testing.T) {
	archive := makeZip(t, map[string]string{
		"LICENSE":   "license",
		"baq.exe":   "WIN-BINARY",
		"README.md": "readme",
	})

	got, err := extractFromZip(archive)
	if err != nil {
		t.Fatalf("extractFromZip: %v", err)
	}
	if string(got) != "WIN-BINARY" {
		t.Errorf("extracted %q, want %q", got, "WIN-BINARY")
	}
}

func TestExtractBinaryDispatch(t *testing.T) {
	zipArchive := makeZip(t, map[string]string{"baq.exe": "win"})
	if got, err := extractBinary("baqueano_Windows_x86_64.zip", zipArchive); err != nil || string(got) != "win" {
		t.Fatalf("extractBinary zip = %q, %v", got, err)
	}

	tarArchive := makeTarGz(t, map[string]string{"baq": "nix"})
	if got, err := extractBinary("baqueano_Linux_x86_64.tar.gz", tarArchive); err != nil || string(got) != "nix" {
		t.Fatalf("extractBinary tar.gz = %q, %v", got, err)
	}
}

func TestExtractBinaryNotFound(t *testing.T) {
	archive := makeZip(t, map[string]string{"README.md": "nope"})
	if _, err := extractBinary("baqueano_Windows_x86_64.zip", archive); err == nil {
		t.Error("expected error when executable is missing, got nil")
	}
}

func TestFindAsset(t *testing.T) {
	release := &version.GitHubRelease{
		TagName: "v1.0.0",
		Assets: []version.GitHubReleaseAsset{
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
			{Name: "baqueano_Linux_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/linux"},
		},
	}

	asset, err := findAsset(release, "BAQUEANO_linux_x86_64.TAR.GZ")
	if err != nil {
		t.Fatalf("findAsset: %v", err)
	}
	if asset.BrowserDownloadURL != "https://example.com/linux" {
		t.Errorf("unexpected asset URL %q", asset.BrowserDownloadURL)
	}

	if _, err := findAsset(release, "missing.tar.gz"); err == nil {
		t.Error("expected error for missing asset, got nil")
	}
}
