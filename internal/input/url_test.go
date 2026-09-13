package input

import (
	"net/url"
	"testing"
)

func TestCleanURL(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantNil    bool
		wantScheme string
	}{
		{
			name:       "http scheme lowercased",
			input:      "HTTP://example.com",
			wantNil:    false,
			wantScheme: "http",
		},
		{
			name:       "https scheme lowercased",
			input:      "HTTPS://example.com",
			wantNil:    false,
			wantScheme: "https",
		},
		{
			name:    "already lowercase http",
			input:   "http://example.com",
			wantNil: false,
			wantScheme: "http",
		},
		{
			name:    "empty string returns non-nil URL",
			input:   "",
			wantNil: false,
		},
		{
			name:    "unparseable control character returns nil",
			input:   "://invalid-url\x7f",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanURL(tt.input)
			if tt.wantNil && got != nil {
				t.Errorf("CleanURL(%q) = non-nil; want nil", tt.input)
				return
			}
			if !tt.wantNil && got == nil {
				t.Errorf("CleanURL(%q) = nil; want non-nil", tt.input)
				return
			}
			if tt.wantScheme != "" && got != nil && got.Scheme != tt.wantScheme {
				t.Errorf("CleanURL(%q).Scheme = %q; want %q", tt.input, got.Scheme, tt.wantScheme)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   *url.URL
		wantErr bool
	}{
		{
			name:    "nil URL panics — not tested here, callers must pass non-nil",
			input:   &url.URL{Scheme: "", Host: ""},
			wantErr: true,
		},
		{
			name:    "missing scheme returns error",
			input:   &url.URL{Scheme: "", Host: "example.com"},
			wantErr: true,
		},
		{
			name:    "ftp scheme returns error",
			input:   &url.URL{Scheme: "ftp", Host: "example.com"},
			wantErr: true,
		},
		{
			name:    "missing host returns error",
			input:   &url.URL{Scheme: "http", Host: ""},
			wantErr: true,
		},
		{
			name:    "valid http URL",
			input:   &url.URL{Scheme: "http", Host: "example.com"},
			wantErr: false,
		},
		{
			name:    "valid https URL",
			input:   &url.URL{Scheme: "https", Host: "example.com"},
			wantErr: false,
		},
		{
			name:    "valid http with port",
			input:   &url.URL{Scheme: "http", Host: "127.0.0.1:9091"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%v) error = %v; wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestCleanThenValidateURL tests the full pipeline as callers use it.
func TestCleanThenValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "valid http", raw: "http://127.0.0.1:9091", wantErr: false},
		{name: "valid https", raw: "https://example.com", wantErr: false},
		{name: "ftp unsupported", raw: "ftp://example.com", wantErr: true},
		{name: "no scheme", raw: "example.com", wantErr: true},
		{name: "missing host", raw: "http://", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := CleanURL(tt.raw)
			if parsed == nil {
				t.Fatalf("CleanURL(%q) returned nil", tt.raw)
			}
			err := ValidateURL(parsed)
			if (err != nil) != tt.wantErr {
				t.Errorf("pipeline(%q) error = %v; wantErr %v", tt.raw, err, tt.wantErr)
			}
		})
	}
}
