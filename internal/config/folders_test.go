package config

import (
	"path/filepath"
	"testing"
)

func TestNormalizeFolderPath(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want string
	}{
		{"", ""},
		{"prod", "prod"},
		{" prod ", "prod"},
		{"prod/web", "prod/web"},
		{"prod / web//db", "prod/web/db"},
		{`prod\web`, "prod/web"},
		{"/prod/", "prod"},
		{"  ", ""},
	} {
		if got := NormalizeFolderPath(tt.in); got != tt.want {
			t.Errorf("NormalizeFolderPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFolderAncestors(t *testing.T) {
	got := FolderAncestors("prod/web/db")
	want := []string{"prod", "prod/web", "prod/web/db"}
	if len(got) != len(want) {
		t.Fatalf("FolderAncestors = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("FolderAncestors[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if FolderAncestors("") != nil {
		t.Error("FolderAncestors(\"\") should be nil")
	}
}

func TestParentFolderAndBase(t *testing.T) {
	if got := ParentFolder("prod/web/db"); got != "prod/web" {
		t.Errorf("ParentFolder = %q, want prod/web", got)
	}
	if got := ParentFolder("prod"); got != "" {
		t.Errorf("ParentFolder(top) = %q, want empty", got)
	}
	if got := FolderBase("prod/web/db"); got != "db" {
		t.Errorf("FolderBase = %q, want db", got)
	}
	if got := FolderDepth("prod/web"); got != 2 {
		t.Errorf("FolderDepth = %d, want 2", got)
	}
	if got := FolderDepth(""); got != 0 {
		t.Errorf("FolderDepth(root) = %d, want 0", got)
	}
}

// newTestStore returns a store backed by a file in a temp directory.
func newTestStore(t *testing.T) *FolderStore {
	t.Helper()
	store, err := LoadFolderStoreFrom(filepath.Join(t.TempDir(), "folders.json"))
	if err != nil {
		t.Fatalf("LoadFolderStoreFrom: %v", err)
	}
	return store
}

func TestFolderStoreCreateAndPersist(t *testing.T) {
	store := newTestStore(t)

	if err := store.CreateFolder("prod/web"); err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	if err := store.SetHostFolder("web1", "prod/web"); err != nil {
		t.Fatalf("SetHostFolder: %v", err)
	}

	// Ancestors are created on demand.
	got := store.AllFolders()
	want := map[string]bool{"prod": true, "prod/web": true}
	if len(got) != len(want) {
		t.Fatalf("AllFolders = %v, want %v", got, want)
	}
	for _, f := range got {
		if !want[f] {
			t.Errorf("unexpected folder %q", f)
		}
	}

	// Reload from disk: the assignment survives.
	reloaded, err := LoadFolderStoreFrom(store.path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := reloaded.FolderOf("web1"); got != "prod/web" {
		t.Errorf("FolderOf(web1) after reload = %q, want prod/web", got)
	}
}

func TestFolderStoreSetHostFolderBackToRoot(t *testing.T) {
	store := newTestStore(t)
	if err := store.SetHostFolder("h", "prod"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetHostFolder("h", ""); err != nil {
		t.Fatal(err)
	}
	if got := store.FolderOf("h"); got != "" {
		t.Errorf("FolderOf = %q, want empty", got)
	}
}

func TestFolderStoreDeleteMovesHostsToParent(t *testing.T) {
	store := newTestStore(t)
	if err := store.CreateFolder("prod/web"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetHostFolder("web1", "prod/web"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetHostFolder("prod1", "prod"); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteFolder("prod/web"); err != nil {
		t.Fatal(err)
	}

	if got := store.FolderOf("web1"); got != "prod" {
		t.Errorf("FolderOf(web1) = %q, want prod (moved to parent)", got)
	}
	if got := store.FolderOf("prod1"); got != "prod" {
		t.Errorf("FolderOf(prod1) = %q, want prod (untouched)", got)
	}
}

func TestFolderStoreDeleteTopLevelMovesHostsToRoot(t *testing.T) {
	store := newTestStore(t)
	if err := store.SetHostFolder("h", "prod"); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteFolder("prod"); err != nil {
		t.Fatal(err)
	}
	if got := store.FolderOf("h"); got != "" {
		t.Errorf("FolderOf(h) = %q, want empty (root)", got)
	}
	if len(store.AllFolders()) != 0 {
		t.Errorf("AllFolders = %v, want empty", store.AllFolders())
	}
}

func TestFolderStoreRenameMovesHostsAndSubfolders(t *testing.T) {
	store := newTestStore(t)
	if err := store.CreateFolder("a/b"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetHostFolder("deep", "a/b"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetHostFolder("shallow", "a"); err != nil {
		t.Fatal(err)
	}

	if err := store.RenameFolder("a", "prod"); err != nil {
		t.Fatal(err)
	}

	if got := store.FolderOf("shallow"); got != "prod" {
		t.Errorf("FolderOf(shallow) = %q, want prod", got)
	}
	if got := store.FolderOf("deep"); got != "prod/b" {
		t.Errorf("FolderOf(deep) = %q, want prod/b", got)
	}
}

func TestFolderStoreIgnoresEmptyNames(t *testing.T) {
	store := newTestStore(t)
	if err := store.CreateFolder("   "); err != nil {
		t.Fatal(err)
	}
	if len(store.AllFolders()) != 0 {
		t.Errorf("AllFolders = %v, want empty", store.AllFolders())
	}
	if err := store.SetHostFolder("  ", "prod"); err != nil {
		t.Fatal(err)
	}
	if len(store.AllFolders()) != 0 {
		t.Errorf("AllFolders = %v, want empty after blank host", store.AllFolders())
	}
}

func TestLoadFolderStoreFromMissingFile(t *testing.T) {
	store, err := LoadFolderStoreFrom(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("LoadFolderStoreFrom missing file: %v", err)
	}
	if len(store.AllFolders()) != 0 || store.FolderOf("anything") != "" {
		t.Error("missing file should yield an empty store")
	}
}
