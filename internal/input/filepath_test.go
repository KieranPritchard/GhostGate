package input

import (
	"testing"
)

func TestCleanFilePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no whitespace",
			input: "/payloads/shell.elf",
			want:  "/payloads/shell.elf",
		},
		{
			name:  "leading and trailing whitespace trimmed",
			input: "  /tmp/x  ",
			want:  "/tmp/x",
		},
		{
			name:  "relative path cleaned",
			input: "payloads/../payloads/shell.elf",
			want:  "payloads/shell.elf",
		},
		{
			name:  "empty string returns dot",
			input: "",
			want:  ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanFilePath(tt.input)
			if got != tt.want {
				t.Errorf("CleanFilePath(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty path returns error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "numeric-only path returns error",
			input:   "12345",
			wantErr: true,
		},
		{
			name:    "special-chars-only path returns error",
			input:   "!@#$%",
			wantErr: true,
		},
		{
			name:    "valid absolute path",
			input:   "/payloads/shell.elf",
			wantErr: false,
		},
		{
			name:    "valid relative path",
			input:   "payloads",
			wantErr: false,
		},
		{
			name:    "path with numbers and letters",
			input:   "/tmp/payload123.bin",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath(%q) error = %v; wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
