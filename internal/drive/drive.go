package drive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

const (
	// DefaultAPIBase is the Drive v3 metadata/media endpoint.
	DefaultAPIBase = "https://www.googleapis.com/drive/v3"
	// DefaultUploadBase is the Drive v3 upload endpoint.
	DefaultUploadBase = "https://www.googleapis.com/upload/drive/v3"
	// AppDataFolder is the hidden per-app folder visible only to this client.
	AppDataFolder = "appDataFolder"
	// DefaultFileName is the document name inside the appDataFolder.
	DefaultFileName = "baqueano_config"
)

// Drive is a Remote backed by a single file in Google Drive's appDataFolder.
// It holds an already-authenticated *http.Client, so it has no OAuth logic of
// its own and can be tested against an httptest server by overriding the bases.
type Drive struct {
	client     *http.Client
	apiBase    string
	uploadBase string
	fileName   string
}

// NewDrive builds a Drive remote around an authenticated HTTP client.
func NewDrive(client *http.Client) *Drive {
	return &Drive{
		client:     client,
		apiBase:    DefaultAPIBase,
		uploadBase: DefaultUploadBase,
		fileName:   DefaultFileName,
	}
}

// fileList is the subset of the Drive files.list response we care about.
type fileList struct {
	Files []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"files"`
}

// findFile returns the document id, or "" when it does not exist yet.
func (d *Drive) findFile() (string, error) {
	q := fmt.Sprintf("name='%s' and '%s' in parents and trashed=false", d.fileName, AppDataFolder)
	endpoint := fmt.Sprintf("%s/files?spaces=%s&pageSize=1&fields=files(id,name)&q=%s",
		d.apiBase, AppDataFolder, url.QueryEscape(q))

	resp, err := d.client.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := ensureOK(resp); err != nil {
		return "", err
	}

	var list fileList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return "", err
	}
	if len(list.Files) == 0 {
		return "", nil
	}
	return list.Files[0].ID, nil
}

// Download implements Remote.
func (d *Drive) Download() ([]byte, error) {
	id, err := d.findFile()
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, nil
	}

	resp, err := d.client.Get(fmt.Sprintf("%s/files/%s?alt=media", d.apiBase, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := ensureOK(resp); err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

// Upload implements Remote: it creates the document on first use and updates it
// afterwards.
func (d *Drive) Upload(data []byte) error {
	id, err := d.findFile()
	if err != nil {
		return err
	}
	if id == "" {
		return d.create(data)
	}
	return d.update(id, data)
}

func (d *Drive) create(data []byte) error {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	metaHeader := textproto.MIMEHeader{}
	metaHeader.Set("Content-Type", "application/json; charset=UTF-8")
	metaPart, err := mw.CreatePart(metaHeader)
	if err != nil {
		return err
	}
	meta, err := json.Marshal(map[string]any{
		"name":    d.fileName,
		"parents": []string{AppDataFolder},
	})
	if err != nil {
		return err
	}
	if _, err := metaPart.Write(meta); err != nil {
		return err
	}

	dataHeader := textproto.MIMEHeader{}
	dataHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	dataPart, err := mw.CreatePart(dataHeader)
	if err != nil {
		return err
	}
	if _, err := dataPart.Write(data); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost,
		d.uploadBase+"/files?uploadType=multipart", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "multipart/related; boundary="+mw.Boundary())
	return d.do(req)
}

func (d *Drive) update(id string, data []byte) error {
	req, err := http.NewRequest(http.MethodPatch,
		fmt.Sprintf("%s/files/%s?uploadType=media", d.uploadBase, id), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
	return d.do(req)
}

func (d *Drive) do(req *http.Request) error {
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return ensureOK(resp)
}

// ensureOK turns a non-2xx response into an error that includes the API message.
func ensureOK(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("drive api %d: %s", resp.StatusCode, msg)
}
