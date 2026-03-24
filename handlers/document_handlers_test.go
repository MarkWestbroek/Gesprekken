package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ──────────────── Mock storage ────────────────

type mockStorage struct {
	data map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{data: make(map[string][]byte)}
}

func (m *mockStorage) Upload(_ context.Context, key string, reader io.Reader, _ int64, _ string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.data[key] = data
	return nil
}

func (m *mockStorage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	data, ok := m.data[key]
	if !ok {
		return nil, fmt.Errorf("object niet gevonden: %s", key)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *mockStorage) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// ──────────────── Helpers ────────────────

func setupDocRouter(dh *DocumentHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/v1")
	v1.POST("/documenten", dh.UploadDocument)
	v1.GET("/documenten/:id", dh.GetDocumentMetadata)
	v1.GET("/documenten/:id/download", dh.DownloadDocument)
	return r
}

// createMultipartBody bouwt een multipart request body met een bestand en form fields.
func createMultipartBody(filename string, content []byte, contentType string, fields map[string]string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if filename != "" {
		h := make(map[string][]string)
		h["Content-Disposition"] = []string{
			fmt.Sprintf(`form-data; name="bestand"; filename="%s"`, filename),
		}
		if contentType != "" {
			h["Content-Type"] = []string{contentType}
		}
		part, _ := writer.CreatePart(h)
		part.Write(content)
	}

	for key, val := range fields {
		writer.WriteField(key, val)
	}
	writer.Close()
	return body, writer.FormDataContentType()
}

// ──────────────── Tests (geen DB - alleen validatie) ────────────────

func TestUpload_MissingFile(t *testing.T) {
	store := newMockStorage()
	// DB is nil — de handler moet falen vóórdat DB wordt aangesproken
	dh := &DocumentHandler{DB: nil, Storage: store}
	r := setupDocRouter(dh)

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	w.WriteField("bronId", uuid.New().String())
	w.Close()

	req := httptest.NewRequest("POST", "/v1/documenten", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("verwacht status 400, kreeg %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), "bestand") {
		t.Errorf("verwacht foutmelding over 'bestand', kreeg: %s", resp.Body.String())
	}
}

func TestUpload_InvalidMIMEType(t *testing.T) {
	store := newMockStorage()
	dh := &DocumentHandler{DB: nil, Storage: store}
	r := setupDocRouter(dh)

	body, ct := createMultipartBody("malware.exe", []byte("bad content"), "application/x-msdownload", map[string]string{
		"bronId": uuid.New().String(),
	})

	req := httptest.NewRequest("POST", "/v1/documenten", body)
	req.Header.Set("Content-Type", ct)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("verwacht status 400, kreeg %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), "niet ondersteund") {
		t.Errorf("verwacht foutmelding over type, kreeg: %s", resp.Body.String())
	}
}

func TestUpload_MissingBronId(t *testing.T) {
	store := newMockStorage()
	dh := &DocumentHandler{DB: nil, Storage: store}
	r := setupDocRouter(dh)

	body, ct := createMultipartBody("test.pdf", []byte("%PDF-1.4 test"), "application/pdf", map[string]string{
		// bronId ontbreekt
	})

	req := httptest.NewRequest("POST", "/v1/documenten", body)
	req.Header.Set("Content-Type", ct)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("verwacht status 400, kreeg %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), "bronId") {
		t.Errorf("verwacht foutmelding over bronId, kreeg: %s", resp.Body.String())
	}
}

func TestDownload_InvalidUUID(t *testing.T) {
	store := newMockStorage()
	dh := &DocumentHandler{DB: nil, Storage: store}
	r := setupDocRouter(dh)

	req := httptest.NewRequest("GET", "/v1/documenten/niet-een-uuid/download", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("verwacht status 400, kreeg %d", resp.Code)
	}
}

func TestMetadata_InvalidUUID(t *testing.T) {
	store := newMockStorage()
	dh := &DocumentHandler{DB: nil, Storage: store}
	r := setupDocRouter(dh)

	req := httptest.NewRequest("GET", "/v1/documenten/niet-een-uuid", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("verwacht status 400, kreeg %d", resp.Code)
	}
}

func TestMimeFromExt(t *testing.T) {
	cases := []struct {
		ext  string
		want string
	}{
		{".pdf", "application/pdf"},
		{".jpg", "image/jpeg"},
		{".png", "image/png"},
		{".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{".xyz", "application/octet-stream"},
	}
	for _, tc := range cases {
		got := mimeFromExt(tc.ext)
		if got != tc.want {
			t.Errorf("mimeFromExt(%q) = %q, wil %q", tc.ext, got, tc.want)
		}
	}
}
