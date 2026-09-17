package input

import (
	"testing"
)

func TestPrepareFilePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "no whitespace",
			input:    "/payloads/shell.elf",
			wantPath: "/payloads/shell.elf",
			wantErr:  false,
		},
		{
			name:     "leading and trailing whitespace trimmed",
			input:    "  /tmp/x  ",
			wantPath: "/tmp/x",
			wantErr:  false,
		},
		{
			name:     "relative path cleaned",
			input:    "payloads/../payloads/shell.elf",
			wantPath: "payloads/shell.elf",
			wantErr:  false,
		},
		{
			name:     "valid absolute path",
			input:    "/payloads/shell.elf",
			wantPath: "/payloads/shell.elf",
			wantErr:  false,
		},
		{
			name:     "valid relative path",
			input:    "payloads",
			wantPath: "payloads",
			wantErr:  false,
		},
		{
			name:     "path with numbers and letters",
			input:    "/tmp/payload123.bin",
			wantPath: "/tmp/payload123.bin",
			wantErr:  false,
		},
		{
			name:     "numeric path (valid filename)",
			input:    "12345",
			wantPath: "12345",
			wantErr:  true,
		},
		{
			name:     "empty string returns error",
			input:    "",
			wantPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := PrepareFilePath(tt.input)

			// Check error presence
			if (err != nil) != tt.wantErr {
				t.Errorf("PrepareFilePath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			// Check output path value if no error was expected
			if !tt.wantErr && path != tt.wantPath {
				t.Errorf("PrepareFilePath(%q) path = %q, wantPath %q", tt.input, path, tt.wantPath)
			}
		})
	}
}