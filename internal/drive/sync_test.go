package drive

import (
	"errors"
	"testing"
)

// fakeRemote is an in-memory Remote for testing the sync engine offline.
type fakeRemote struct {
	data       []byte
	uploaded   []byte
	uploadedN  int
	downloadFn func() ([]byte, error)
	uploadFn   func([]byte) error
}

func (f *fakeRemote) Download() ([]byte, error) {
	if f.downloadFn != nil {
		return f.downloadFn()
	}
	return f.data, nil
}

func (f *fakeRemote) Upload(data []byte) error {
	if f.uploadFn != nil {
		return f.uploadFn(data)
	}
	f.uploaded = data
	f.uploadedN++
	f.data = data
	return nil
}

func TestSyncMergesAndUploads(t *testing.T) {
	remote := &fakeRemote{data: []byte(remoteConfig)}
	res, err := Sync(remote, []byte(localConfig))
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if res.RemoteEmpty {
		t.Error("RemoteEmpty = true, want false")
	}
	if res.Stats.Added != 1 {
		t.Errorf("Added = %d, want 1", res.Stats.Added)
	}
	if string(remote.uploaded) != string(res.Content) {
		t.Error("uploaded content differs from returned content")
	}
	if !containsLine(string(remote.uploaded), "Host gamma") {
		t.Errorf("remote-only host not uploaded:\n%s", remote.uploaded)
	}
}

func TestSyncRemoteEmpty(t *testing.T) {
	remote := &fakeRemote{}
	res, err := Sync(remote, []byte(localConfig))
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if !res.RemoteEmpty {
		t.Error("RemoteEmpty = false, want true")
	}
	if res.Stats.Added != 0 {
		t.Errorf("Added = %d, want 0", res.Stats.Added)
	}
}

func TestSyncDownloadError(t *testing.T) {
	remote := &fakeRemote{downloadFn: func() ([]byte, error) {
		return nil, errors.New("boom")
	}}
	if _, err := Sync(remote, []byte(localConfig)); err == nil {
		t.Fatal("Sync error = nil, want download failure")
	}
}

func TestSyncUploadError(t *testing.T) {
	remote := &fakeRemote{
		data:     []byte(remoteConfig),
		uploadFn: func([]byte) error { return errors.New("disk full") },
	}
	if _, err := Sync(remote, []byte(localConfig)); err == nil {
		t.Fatal("Sync error = nil, want upload failure")
	}
}

func TestPushOverwritesWithoutMerge(t *testing.T) {
	remote := &fakeRemote{data: []byte(remoteConfig)}
	if err := Push(remote, []byte(localConfig)); err != nil {
		t.Fatalf("Push error: %v", err)
	}
	if string(remote.uploaded) != localConfig {
		t.Errorf("Push uploaded %q, want the raw local config", remote.uploaded)
	}
	if containsLine(string(remote.uploaded), "Host gamma") {
		t.Error("Push must not merge remote-only hosts")
	}
}

func TestPullReturnsRemote(t *testing.T) {
	remote := &fakeRemote{data: []byte(remoteConfig)}
	data, err := Pull(remote)
	if err != nil {
		t.Fatalf("Pull error: %v", err)
	}
	if string(data) != remoteConfig {
		t.Errorf("Pull = %q, want the remote config", data)
	}
}
