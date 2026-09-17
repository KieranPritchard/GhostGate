package input

import (
	"testing"
)

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
			// Prepares the target
			target, err := PrepareURL(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrepareURL(%q) error = %v (target = %q), wantErr %v", tt.raw, err, target, tt.wantErr)
			}
		})
	}
}
