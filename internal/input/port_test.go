package input

import (
	"testing"
)

func TestPreparePort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "no whitespace",
			input:   "8080",
			wantErr: false,
		},
		{
			name:    "leading and trailing whitespace trimmed",
			input:   "  8080  ",
			wantErr: false,
		},
		{
			name:    "only whitespace becomes empty",
			input:   "   ",
			wantErr: true,
		},
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
			port, err := PreparePort(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%q) error = %v on target = %s; wantErr %v", tt.input, err, port, tt.wantErr)
			}
		})
	}
}
