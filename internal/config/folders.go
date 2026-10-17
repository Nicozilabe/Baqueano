package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// FolderStore keeps the user's folder organization: which folder each host
// belongs to, plus the set of folders that were explicitly created (so empty
// folders survive restarts).
//
// It is stored in its own JSON file inside the baqueano config directory so the
// user's ~/.ssh/config is never touched by the folder feature.
type FolderStore struct {
	// Folders is the set of explicitly created folder paths, e.g. "prod" or
	// "prod/web".
	Folders []string `json:"folders"`
	// Hosts maps a host alias to its folder path. An empty string (or a missing
	// entry) means the host lives at the root.
	Hosts map[string]string `json:"hosts"`

	path string
	mu   sync.Mutex
}

// NormalizeFolderPath cleans a folder path so that the same folder typed in
// different ways ends up identical: no surrounding spaces, forward slashes
// only, no empty or duplicated segments. " prod / web//db " becomes
// "prod/web/db".
func NormalizeFolderPath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	segments := strings.Split(path, "/")
	clean := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		clean = append(clean, segment)
	}
	return strings.Join(clean, "/")
}

// ParentFolder returns the parent of a folder path, or "" for a top-level
// folder.
func ParentFolder(path string) string {
	path = NormalizeFolderPath(path)
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return ""
	}
	return path[:idx]
}

// FolderBase returns the last segment of a folder path.
func FolderBase(path string) string {
	path = NormalizeFolderPath(path)
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return path
	}
	return path[idx+1:]
}

// FolderAncestors returns every ancestor of a folder path, from the top-level
// folder down to the path itself. "prod/web/db" yields
// ["prod", "prod/web", "prod/web/db"].
func FolderAncestors(path string) []string {
	path = NormalizeFolderPath(path)
	if path == "" {
		return nil
	}
	segments := strings.Split(path, "/")
	ancestors := make([]string, 0, len(segments))
	current := ""
	for _, segment := range segments {
		if current == "" {
			current = segment
		} else {
			current = current + "/" + segment
		}
		ancestors = append(ancestors, current)
	}
	return ancestors
}

// FolderDepth returns how deep a folder path is, counting from 1. The root
// folder ("") is depth 0.
func FolderDepth(path string) int {
	path = NormalizeFolderPath(path)
	if path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}

// GetFolderStorePath returns the path of the folder store file.
func GetFolderStorePath() (string, error) {
	configDir, err := GetBaqueanoConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "folders.json"), nil
}

// LoadFolderStore reads the folder organization from disk. A missing file is
// not an error: it simply means no folders have been created yet.
func LoadFolderStore() (*FolderStore, error) {
	path, err := GetFolderStorePath()
	if err != nil {
		return nil, err
	}
	return LoadFolderStoreFrom(path)
}

// LoadFolderStoreFrom reads the folder organization from an explicit path.
func LoadFolderStoreFrom(path string) (*FolderStore, error) {
	store := &FolderStore{
		Folders: nil,
		Hosts:   map[string]string{},
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}

	var persisted FolderStore
	if err := json.Unmarshal(data, &persisted); err != nil {
		return nil, err
	}

	if persisted.Hosts != nil {
		store.Hosts = persisted.Hosts
	}
	store.Folders = dedupeFolders(persisted.Folders)

	return store, nil
}

// dedupeFolders normalizes and deduplicates a list of folder paths, expanding
// ancestor folders so a nested path always has its parents.
func dedupeFolders(folders []string) []string {
	seen := map[string]bool{}
	var result []string
	add := func(path string) {
		path = NormalizeFolderPath(path)
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		result = append(result, path)
	}
	for _, folder := range folders {
		for _, ancestor := range FolderAncestors(folder) {
			add(ancestor)
		}
	}
	sort.Strings(result)
	return result
}

// Save writes the folder organization back to disk.
func (s *FolderStore) Save() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *FolderStore) saveLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	s.Folders = dedupeFolders(s.Folders)
	sort.Strings(s.Folders)
	if s.Hosts == nil {
		s.Hosts = map[string]string{}
	}

	// Drop empty assignments so the file stays tidy.
	for host, folder := range s.Hosts {
		if NormalizeFolderPath(folder) == "" {
			delete(s.Hosts, host)
		}
	}

	// Forget folder assignments that no longer have a folder.
	for host, folder := range s.Hosts {
		if !containsFolder(s.Folders, folder) {
			delete(s.Hosts, host)
		}
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func containsFolder(folders []string, folder string) bool {
	for _, f := range folders {
		if f == folder {
			return true
		}
	}
	return false
}

// FolderOf returns the folder a host belongs to, or "" when it lives at the
// root.
func (s *FolderStore) FolderOf(hostName string) string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return NormalizeFolderPath(s.Hosts[hostName])
}

// AllFolders returns a copy of the explicit folder list.
func (s *FolderStore) AllFolders() []string {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]string, len(s.Folders))
	copy(result, s.Folders)
	return result
}

// CreateFolder registers a folder (and its ancestors) as existing.
func (s *FolderStore) CreateFolder(path string) error {
	if s == nil {
		return nil
	}
	path = NormalizeFolderPath(path)
	if path == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Folders = dedupeFolders(append(s.Folders, path))
	return s.saveLocked()
}

// SetHostFolder moves a host into a folder. An empty folder moves the host back
// to the root. The target folder and its ancestors are created on demand.
func (s *FolderStore) SetHostFolder(hostName, folder string) error {
	if s == nil {
		return nil
	}
	hostName = strings.TrimSpace(hostName)
	if hostName == "" {
		return nil
	}
	folder = NormalizeFolderPath(folder)

	s.mu.Lock()
	defer s.mu.Unlock()

	if folder == "" {
		delete(s.Hosts, hostName)
	} else {
		if s.Hosts == nil {
			s.Hosts = map[string]string{}
		}
		s.Hosts[hostName] = folder
		s.Folders = dedupeFolders(append(s.Folders, folder))
	}
	return s.saveLocked()
}

// DeleteFolder removes a folder and all of its subfolders. Hosts that lived in
// the deleted folders are moved to the deleted folder's parent.
func (s *FolderStore) DeleteFolder(path string) error {
	if s == nil {
		return nil
	}
	path = NormalizeFolderPath(path)
	if path == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	parent := ParentFolder(path)
	prefix := path + "/"

	remaining := make([]string, 0, len(s.Folders))
	for _, folder := range s.Folders {
		if folder == path || strings.HasPrefix(folder, prefix) {
			continue
		}
		remaining = append(remaining, folder)
	}
	s.Folders = remaining

	for host, folder := range s.Hosts {
		if folder == path || strings.HasPrefix(folder, prefix) {
			if parent == "" {
				delete(s.Hosts, host)
			} else {
				s.Hosts[host] = parent
			}
		}
	}

	return s.saveLocked()
}

// RenameFolder moves a folder (and its descendants) to a new path, updating the
// hosts and subfolders that lived inside it.
func (s *FolderStore) RenameFolder(oldPath, newPath string) error {
	if s == nil {
		return nil
	}
	oldPath = NormalizeFolderPath(oldPath)
	newPath = NormalizeFolderPath(newPath)
	if oldPath == "" || newPath == "" || oldPath == newPath {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldPrefix := oldPath + "/"
	for i, folder := range s.Folders {
		switch {
		case folder == oldPath:
			s.Folders[i] = newPath
		case strings.HasPrefix(folder, oldPrefix):
			s.Folders[i] = newPath + folder[len(oldPath):]
		}
	}

	for host, folder := range s.Hosts {
		switch {
		case folder == oldPath:
			s.Hosts[host] = newPath
		case strings.HasPrefix(folder, oldPrefix):
			s.Hosts[host] = newPath + folder[len(oldPath):]
		}
	}

	s.Folders = dedupeFolders(append(s.Folders, newPath))
	return s.saveLocked()
}
