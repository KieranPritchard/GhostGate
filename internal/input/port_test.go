package input

import (
	"testing"
)

func TestCleanPort(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no whitespace",
			input: "8080",
			want:  "8080",
		},
		{
			name:  "leading and trailing whitespace trimmed",
			input: "  8080  ",
			want:  "8080",
		},
		{
			name:  "only whitespace becomes empty",
			input: "   ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanPort(tt.input)
			if got != tt.want {
				t.Errorf("CleanPort(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "port too high (999999)",
			input:   "999999",
			wantErr: true,
		},
		{
			name:    "port zero",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "negative port",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "non-numeric string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "valid lower bound (1)",
			input:   "1",
			wantErr: false,
		},
		{
			name:    "valid upper bound (65535)",
			input:   "65535",
			wantErr: false,
		},
		{
			name:    "valid common port (8080)",
			input:   "8080",
			wantErr: false,
		},
		{
			name:    "valid HTTPS port (443)",
			input:   "443",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%q) error = %v; wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
