package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandleTunnel_RelaysResponseBody verifies the tunnel proxies the response body correctly.
func TestHandleTunnel_RelaysResponseBody(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	handler := HandleTunnel(backend.URL)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "backend response" {
		t.Errorf("body = %q; want %q", body, "backend response")
	}
}

// TestHandleTunnel_ForwardsRequestHeaders verifies that incoming headers are forwarded to the backend.
func TestHandleTunnel_ForwardsRequestHeaders(t *testing.T) {
	var receivedHeader string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Header")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := HandleTunnel(backend.URL)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom-Header", "ghostgate-test")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if receivedHeader != "ghostgate-test" {
		t.Errorf("backend received header = %q; want %q", receivedHeader, "ghostgate-test")
	}
}

// TestHandleTunnel_RelaysResponseHeaders verifies that backend response headers are relayed to the caller.
func TestHandleTunnel_RelaysResponseHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend-Relay", "relay-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := HandleTunnel(backend.URL)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Header().Get("X-Backend-Relay") != "relay-value" {
		t.Errorf("relayed header X-Backend-Relay = %q; want %q",
			rec.Header().Get("X-Backend-Relay"), "relay-value")
	}
}

// TestHandleTunnel_BadUpstreamReturns502 verifies that an unreachable upstream yields 502.
func TestHandleTunnel_BadUpstreamReturns502(t *testing.T) {
	// Port 1 is reserved and will always fail to connect.
	handler := HandleTunnel("http://127.0.0.1:1")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status %d for bad upstream; got %d", http.StatusBadGateway, rec.Code)
	}
}

// TestHandleTunnel_StatusCodeRelayed verifies non-200 backend status codes are relayed through.
func TestHandleTunnel_StatusCodeRelayed(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer backend.Close()

	handler := HandleTunnel(backend.URL)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d; got %d", http.StatusNotFound, rec.Code)
	}
}

// TestHandleTunnel_POSTBodyRelayed verifies POST bodies are forwarded to the backend.
func TestHandleTunnel_POSTBodyRelayed(t *testing.T) {
	const payload = "post payload content"
	var receivedBody string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		receivedBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := HandleTunnel(backend.URL)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if receivedBody != payload {
		t.Errorf("backend received body = %q; want %q", receivedBody, payload)
	}
}

// TestHandleTunnel_InvalidTarget verifies 500 when target URL is malformed for NewRequest.
func TestHandleTunnel_InvalidTarget(t *testing.T) {
	handler := HandleTunnel("://bad-target\x7f")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for invalid target URL; got %d", http.StatusInternalServerError, rec.Code)
	}
}

