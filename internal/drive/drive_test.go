package drive

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// driveStub is a minimal in-memory Google Drive API used to exercise the real
// HTTP code paths without touching the network.
type driveStub struct {
	fileID   string
	data     []byte
	lastMeta map[string]any
	created  bool
	updated  bool
}

func (s *driveStub) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/files":
			q := r.URL.Query().Get("q")
			if !strings.Contains(q, "appDataFolder") {
				t := "bad query: " + q
				http.Error(w, t, http.StatusBadRequest)
				return
			}
			if s.fileID == "" {
				io.WriteString(w, `{"files":[]}`)
				return
			}
			io.WriteString(w, `{"files":[{"id":"`+s.fileID+`","name":"baqueano_config"}]}`)

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/files/"):
			if r.URL.Query().Get("alt") != "media" {
				http.Error(w, "expected alt=media", http.StatusBadRequest)
				return
			}
			w.Write(s.data)

		case r.Method == http.MethodPost && r.URL.Path == "/files":
			if r.URL.Query().Get("uploadType") != "multipart" {
				http.Error(w, "expected uploadType=multipart", http.StatusBadRequest)
				return
			}
			mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
				http.Error(w, "expected multipart body", http.StatusBadRequest)
				return
			}
			parts := readMultipart(t, r.Body, params["boundary"])
			if len(parts) != 2 {
				http.Error(w, "expected 2 parts", http.StatusBadRequest)
				return
			}
			s.lastMeta = parseJSONMap(t, parts[0])
			s.data = parts[1]
			s.fileID = "created-id"
			s.created = true
			io.WriteString(w, `{"id":"created-id"}`)

		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/files/"):
			if r.URL.Query().Get("uploadType") != "media" {
				http.Error(w, "expected uploadType=media", http.StatusBadRequest)
				return
			}
			body, _ := io.ReadAll(r.Body)
			s.data = body
			s.updated = true
			id := strings.TrimPrefix(r.URL.Path, "/files/")
			io.WriteString(w, `{"id":"`+id+`"}`)

		default:
			http.NotFound(w, r)
		}
	})
}

func newDriveForStub(t *testing.T, s *driveStub) (*Drive, *httptest.Server) {
	srv := httptest.NewServer(s.handler(t))
	d := NewDrive(srv.Client())
	d.apiBase = srv.URL
	d.uploadBase = srv.URL
	return d, srv
}

func TestDriveDownloadNoFile(t *testing.T) {
	stub := &driveStub{}
	d, srv := newDriveForStub(t, stub)
	defer srv.Close()

	data, err := d.Download()
	if err != nil {
		t.Fatalf("Download error: %v", err)
	}
	if data != nil {
		t.Errorf("Download = %q, want nil for a missing file", data)
	}
}

func TestDriveDownloadExisting(t *testing.T) {
	stub := &driveStub{fileID: "abc", data: []byte("Host a\n")}
	d, srv := newDriveForStub(t, stub)
	defer srv.Close()

	data, err := d.Download()
	if err != nil {
		t.Fatalf("Download error: %v", err)
	}
	if string(data) != "Host a\n" {
		t.Errorf("Download = %q, want the stored data", data)
	}
}

func TestDriveUploadCreatesFile(t *testing.T) {
	stub := &driveStub{}
	d, srv := newDriveForStub(t, stub)
	defer srv.Close()

	if err := d.Upload([]byte("Host new\n")); err != nil {
		t.Fatalf("Upload error: %v", err)
	}
	if !stub.created {
		t.Fatal("expected a multipart create")
	}
	if got := stub.lastMeta["name"]; got != DefaultFileName {
		t.Errorf("metadata name = %v, want %q", got, DefaultFileName)
	}
	parents, _ := stub.lastMeta["parents"].([]any)
	if len(parents) != 1 || parents[0] != AppDataFolder {
		t.Errorf("metadata parents = %v, want [%s]", stub.lastMeta["parents"], AppDataFolder)
	}
	if string(stub.data) != "Host new\n" {
		t.Errorf("uploaded data = %q", stub.data)
	}
}

func TestDriveUploadUpdatesFile(t *testing.T) {
	stub := &driveStub{fileID: "abc", data: []byte("old\n")}
	d, srv := newDriveForStub(t, stub)
	defer srv.Close()

	if err := d.Upload([]byte("Host new\n")); err != nil {
		t.Fatalf("Upload error: %v", err)
	}
	if !stub.updated {
		t.Fatal("expected a media update")
	}
	if stub.created {
		t.Error("update must not create a new file")
	}
	if string(stub.data) != "Host new\n" {
		t.Errorf("updated data = %q", stub.data)
	}
}

func TestDriveUploadReportsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "insufficientPermissions", http.StatusForbidden)
	}))
	defer srv.Close()

	d := NewDrive(srv.Client())
	d.apiBase = srv.URL
	d.uploadBase = srv.URL

	err := d.Upload([]byte("Host new\n"))
	if err == nil {
		t.Fatal("Upload error = nil, want API error")
	}
	if !strings.Contains(err.Error(), "drive api 403") ||
		!strings.Contains(err.Error(), "insufficientPermissions") {
		t.Errorf("error = %v, want it to include the API message", err)
	}
}

// readMultipart returns the raw bodies of the multipart parts, in order. It
// parses multipart/related manually because http.Request.MultipartReader only
// accepts multipart/form-data or multipart/mixed.
func readMultipart(t *testing.T, body io.Reader, boundary string) [][]byte {
	t.Helper()
	mr := multipart.NewReader(body, boundary)
	var parts [][]byte
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("NextPart: %v", err)
		}
		part, _ := io.ReadAll(p)
		parts = append(parts, part)
	}
	return parts
}

func parseJSONMap(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal metadata: %v (%s)", err, b)
	}
	return m
}
