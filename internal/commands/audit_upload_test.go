package commands

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestFormatStatus verifies that status codes are coloured correctly.
func TestFormatStatus(t *testing.T) {
	tests := []struct {
		status      string
		wantContain string
	}{
		{"200 OK", ColorGreen},
		{"201 Created", ColorGreen},
		{"404 Not Found", ColorRed},
		{"500 Internal Server Error", ColorRed},
		{"302 Found", ColorYellow},
		{"101 Switching Protocols", ColorYellow},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := formatStatus(tt.status)
			if len(got) == 0 {
				t.Errorf("formatStatus(%q) returned empty string", tt.status)
			}
			if !bytes.Contains([]byte(got), []byte(tt.wantContain)) {
				t.Errorf("formatStatus(%q) = %q; expected to contain ANSI code %q", tt.status, got, tt.wantContain)
			}
		})
	}
}

// TestIsSelfClosing checks known self-closing tags and a regular tag.
func TestIsSelfClosing(t *testing.T) {
	selfClosing := []string{"meta", "link", "br", "hr", "img", "input"}
	for _, tag := range selfClosing {
		t.Run(tag, func(t *testing.T) {
			if !isSelfClosing(tag) {
				t.Errorf("isSelfClosing(%q) = false; want true", tag)
			}
		})
	}

	notSelfClosing := []string{"div", "span", "p", "a", "body", "html", "ul"}
	for _, tag := range notSelfClosing {
		t.Run(tag, func(t *testing.T) {
			if isSelfClosing(tag) {
				t.Errorf("isSelfClosing(%q) = true; want false", tag)
			}
		})
	}
}

// TestAuditRequest_Nil ensures AuditRequest does not panic on a nil response.
func TestAuditRequest_Nil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AuditRequest(nil) panicked: %v", r)
		}
	}()
	// Redirect stdout so the test output stays clean.
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	AuditRequest(nil)
}

// TestAuditRequest_JSONBody verifies that a JSON response body is handled without panic.
func TestAuditRequest_JSONBody(t *testing.T) {
	body := `{"key":"value","count":42}`
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusOK)
	rec.Body.WriteString(body)

	resp := rec.Result()

	// Redirect stdout to a buffer so we can assert on it.
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	AuditRequest(resp)

	w.Close()
	os.Stdout = old

	var out bytes.Buffer
	io.Copy(&out, r)

	if !bytes.Contains(out.Bytes(), []byte("TARGET HTTP SERVER AUDIT")) {
		t.Errorf("expected audit header in output; got:\n%s", out.String())
	}
}

// TestAuditRequest_HTMLBody verifies that an HTML response body is handled without panic.
func TestAuditRequest_HTMLBody(t *testing.T) {
	body := `<html><head><title>Test</title></head><body><p>Hello</p></body></html>`
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "text/html; charset=utf-8")
	rec.WriteHeader(http.StatusOK)
	rec.Body.WriteString(body)

	resp := rec.Result()

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	AuditRequest(resp)

	w.Close()
	os.Stdout = old

	var out bytes.Buffer
	io.Copy(&out, r)

	if !bytes.Contains(out.Bytes(), []byte("BODY")) {
		t.Errorf("expected BODY section in HTML audit output; got:\n%s", out.String())
	}
}

// TestAuditRequest_PlainTextBody verifies plain text content-type is handled.
func TestAuditRequest_PlainTextBody(t *testing.T) {
	body := "plain text response body"
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "text/plain")
	rec.WriteHeader(http.StatusOK)
	rec.Body.WriteString(body)

	resp := rec.Result()

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	AuditRequest(resp)

	w.Close()
	os.Stdout = old

	var out bytes.Buffer
	io.Copy(&out, r)

	if !bytes.Contains(out.Bytes(), []byte(body)) {
		t.Errorf("expected plain body %q in output; got:\n%s", body, out.String())
	}
}

// TestPrettyPrintHTML verifies that prettyPrintHTML indents HTML correctly and includes key tags.
func TestPrettyPrintHTML(t *testing.T) {
	input := `<html><body><p>Hello</p></body></html>`
	output := prettyPrintHTML(input)

	if len(output) == 0 {
		t.Error("prettyPrintHTML() returned empty string")
	}
	if !bytes.Contains([]byte(output), []byte("<html>")) {
		t.Errorf("expected <html> in output; got:\n%s", output)
	}
	if !bytes.Contains([]byte(output), []byte("Hello")) {
		t.Errorf("expected text content in output; got:\n%s", output)
	}
}

// TestPrettyPrintHTML_SelfClosingTags ensures self-closing tags don't increase indent forever.
func TestPrettyPrintHTML_SelfClosingTags(t *testing.T) {
	input := `<html><head><meta charset="utf-8"><link rel="stylesheet" href="style.css"></head><body><img src="x.png"><br></body></html>`
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("prettyPrintHTML panicked on self-closing tags: %v", r)
		}
	}()
	_ = prettyPrintHTML(input)
}

// TestAuditRequest_EmptyBody verifies a response with no body is handled gracefully.
func TestAuditRequest_EmptyBody(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusNoContent)
	resp := rec.Result()

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AuditRequest panicked on empty body: %v", r)
		}
	}()
	AuditRequest(resp)

	w.Close()
	os.Stdout = old

	var out bytes.Buffer
	io.Copy(&out, r)
	_ = out
}

// TestAuditRequest_LiveServer runs AuditRequest against a real httptest.Server.
func TestAuditRequest_LiveServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET %s failed: %v", ts.URL, err)
	}

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	AuditRequest(resp)

	w.Close()
	os.Stdout = old

	var out bytes.Buffer
	io.Copy(&out, r)

	if !bytes.Contains(out.Bytes(), []byte("status")) {
		t.Errorf("expected JSON field in audit output; got:\n%s", out.String())
	}
}

// helper to create a temp dir with a file for upload tests.
func tempExfilDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// TestUploadHandler_POST verifies a POST with X-File-Name header saves file and returns 201.
func TestUploadHandler_POST(t *testing.T) {
	exfilDir := tempExfilDir(t)
	handler := UploadHandler(exfilDir)

	body := bytes.NewBufferString("exfiltrated secret data")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("X-File-Name", "secret.txt")

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d; got %d", http.StatusCreated, rec.Code)
	}

	savedPath := filepath.Join(exfilDir, "secret.txt")
	content, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("saved file not found at %s: %v", savedPath, err)
	}
	if string(content) != "exfiltrated secret data" {
		t.Errorf("saved content = %q; want %q", content, "exfiltrated secret data")
	}
}

// TestUploadHandler_POST_NoFilename defaults to exfil_data.bin.
func TestUploadHandler_POST_NoFilename(t *testing.T) {
	exfilDir := tempExfilDir(t)
	handler := UploadHandler(exfilDir)

	body := bytes.NewBufferString("default filename data")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	// No X-File-Name header.

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d; got %d", http.StatusCreated, rec.Code)
	}

	savedPath := filepath.Join(exfilDir, "exfil_data.bin")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Errorf("expected default file %s to exist, but it doesn't", savedPath)
	}
}

// TestUploadHandler_NonPOST verifies that non-POST methods return 405.
func TestUploadHandler_NonPOST(t *testing.T) {
	exfilDir := tempExfilDir(t)
	handler := UploadHandler(exfilDir)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/upload", nil)
		rec := httptest.NewRecorder()
		handler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("[%s] expected status %d; got %d", method, http.StatusMethodNotAllowed, rec.Code)
		}
	}
}

// TestUploadHandler_PathTraversalPrevented verifies path traversal in filename is sanitised.
func TestUploadHandler_PathTraversalPrevented(t *testing.T) {
	exfilDir := tempExfilDir(t)
	handler := UploadHandler(exfilDir)

	body := bytes.NewBufferString("malicious content")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("X-File-Name", "../../etc/passwd")

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201 even for traversal attempt; got %d", rec.Code)
	}

	// The file must be saved inside exfilDir, not outside it.
	// filepath.Base("../../etc/passwd") == "passwd"
	savedPath := filepath.Join(exfilDir, "passwd")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Error("expected sanitised file to be saved inside exfilDir")
	}
}

// errReader produces an error when read.
type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

// TestUploadHandler_MkdirError verifies 500 when exfilDir cannot be created.
func TestUploadHandler_MkdirError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file_blocking_dir")
	_ = os.WriteFile(tmpFile, []byte("x"), 0644)

	// Passing a path where a regular file exists as a directory component
	blockedDir := filepath.Join(tmpFile, "cannot_create_dir")
	handler := UploadHandler(blockedDir)

	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("test"))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when directory cannot be created, got %d", rec.Code)
	}
}

// TestUploadHandler_BodyReadError verifies 500 when reading the request body fails.
func TestUploadHandler_BodyReadError(t *testing.T) {
	exfilDir := tempExfilDir(t)
	handler := UploadHandler(exfilDir)

	req := httptest.NewRequest(http.MethodPost, "/upload", &errReader{})
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on body read failure, got %d", rec.Code)
	}
}

