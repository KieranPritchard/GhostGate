package commands

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestCheckInsecureHeader covers the per-header warning logic
func TestCheckInsecureHeader(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		wantSubstr  string // substring expected in the output, empty means no warning
		wantNoWarn  bool
	}{
		{
			name:       "server header discloses tech stack",
			key:        "Server",
			value:      "nginx/1.18.0",
			wantSubstr: "Information disclosure",
		},
		{
			name:       "x-powered-by discloses tech stack",
			key:        "X-Powered-By",
			value:      "PHP/7.4.3",
			wantSubstr: "Information disclosure",
		},
		{
			name:       "wildcard CORS",
			key:        "Access-Control-Allow-Origin",
			value:      "*",
			wantSubstr: "Wildcard CORS",
		},
		{
			name:       "non-wildcard CORS is not flagged",
			key:        "Access-Control-Allow-Origin",
			value:      "https://example.com",
			wantNoWarn: true,
		},
		{
			name:       "cookie missing all flags",
			key:        "Set-Cookie",
			value:      "session=abc123",
			wantSubstr: "missing Secure flag",
		},
		{
			name:       "cookie with all flags set is not flagged",
			key:        "Set-Cookie",
			value:      "session=abc123; Secure; HttpOnly; SameSite=Strict",
			wantNoWarn: true,
		},
		{
			name:       "xss protection disabled",
			key:        "X-XSS-Protection",
			value:      "0",
			wantSubstr: "XSS auditor disabled",
		},
		{
			name:       "xss protection enabled is still deprecated",
			key:        "X-XSS-Protection",
			value:      "1; mode=block",
			wantSubstr: "Deprecated header",
		},
		{
			name:       "csp with unsafe-inline",
			key:        "Content-Security-Policy",
			value:      "default-src 'self' 'unsafe-inline'",
			wantSubstr: "Weak CSP policy",
		},
		{
			name:       "strict csp is not flagged",
			key:        "Content-Security-Policy",
			value:      "default-src 'self'",
			wantNoWarn: true,
		},
		{
			name:       "unrelated header is not flagged",
			key:        "Content-Type",
			value:      "application/json",
			wantNoWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkInsecureHeader(tt.key, tt.value)

			// Checks that headers we expect to be clean produce no warning
			if tt.wantNoWarn && got != "" {
				t.Errorf("checkInsecureHeader(%q, %q) = %q, want no warning", tt.key, tt.value, got)
				return
			}

			// Checks that the expected warning text is present
			if tt.wantSubstr != "" && !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("checkInsecureHeader(%q, %q) = %q, want substring %q", tt.key, tt.value, got, tt.wantSubstr)
			}
		})
	}
}

// TestScanResponseBody covers secret detection across the default rule set
func TestScanResponseBody(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantType  string
		wantCount int
	}{
		{
			name:      "aws access key detected",
			body:      "config: " + "AKIA" + "ABCDEFGHIJKLMNOP",
			wantType:  "AWS Access Key ID",
			wantCount: 1,
		},
		{
			name:      "github token detected",
			body:      "token=" + "ghp_" + "1234567890abcdef1234567890abcdef1234",
			wantType:  "GitHub Personal Access Token",
			wantCount: 1,
		},
		{
			name:      "private key detected",
			body:      "-----BEGIN RSA PRIV" + "ATE KEY-----\nMIIB...\n-----END RSA PRIVATE KEY-----",
			wantType:  "Private Key",
			wantCount: 1,
		},
		{
			name:      "clean body has no findings",
			body:      `{"status":"ok","message":"nothing to see here"}`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := ScanResponseBody([]byte(tt.body))

			if len(findings) != tt.wantCount {
				t.Fatalf("ScanResponseBody(%q) returned %d findings, want %d", tt.body, len(findings), tt.wantCount)
			}

			// Skips the type check when no findings were expected
			if tt.wantCount == 0 {
				return
			}

			found := false
			for _, f := range findings {
				if f.Type == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ScanResponseBody(%q) did not include expected type %q", tt.body, tt.wantType)
			}
		})
	}
}

// TestFormatFindings covers empty input and truncation of long matches
func TestFormatFindings(t *testing.T) {
	tests := []struct {
		name       string
		findings   []Finding
		wantEmpty  bool
		wantSubstr string
	}{
		{
			name:      "no findings returns empty string",
			findings:  nil,
			wantEmpty: true,
		},
		{
			name: "short match is printed in full",
			findings: []Finding{
				{Type: "Test Secret", Match: "short-value"},
			},
			wantSubstr: "short-value",
		},
		{
			name: "long match is truncated to 60 chars with ellipsis",
			findings: []Finding{
				{Type: "Test Secret", Match: strings.Repeat("a", 100)},
			},
			wantSubstr: strings.Repeat("a", 57) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFindings(tt.findings)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("FormatFindings() = %q, want empty string", got)
				}
				return
			}

			if !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("FormatFindings() = %q, want substring %q", got, tt.wantSubstr)
			}
		})
	}
}

// TestAuditHeaders covers missing-header detection and insecure-value flagging
func TestAuditHeaders(t *testing.T) {
	tests := []struct {
		name        string
		header      http.Header
		wantSubstrs []string // all of these must appear in the output
	}{
		{
			name:   "no security headers present, all reported missing",
			header: http.Header{"Content-Type": []string{"text/plain"}},
			wantSubstrs: []string{
				"Strict-Transport-Security",
				"Content-Security-Policy",
				"X-Frame-Options",
				"X-Content-Type-Options",
				"Referrer-Policy",
				"Permissions-Policy",
			},
		},
		{
			name: "all security headers present, none reported missing",
			header: http.Header{
				"Strict-Transport-Security": []string{"max-age=63072000"},
				"Content-Security-Policy":   []string{"default-src 'self'"},
				"X-Frame-Options":           []string{"DENY"},
				"X-Content-Type-Options":    []string{"nosniff"},
				"Referrer-Policy":           []string{"no-referrer"},
				"Permissions-Policy":        []string{"geolocation=()"},
			},
		},
		{
			name:   "insecure header value is flagged inline",
			header: http.Header{"Server": []string{"Apache/2.4.1"}},
			wantSubstrs: []string{
				"Information disclosure",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			AuditHeaders(&buf, tt.header)
			got := buf.String()

			for _, want := range tt.wantSubstrs {
				if !strings.Contains(got, want) {
					t.Errorf("AuditHeaders() output missing %q\nfull output:\n%s", want, got)
				}
			}

			// Checks that the "all present" case does not print a missing-headers section
			if tt.name == "all security headers present, none reported missing" && strings.Contains(got, "MISSING SECURITY HEADERS") {
				t.Errorf("AuditHeaders() unexpectedly reported missing headers:\n%s", got)
			}
		})
	}
}

// TestExtractHTMLSecurityHits covers title, form, input, link, script and comment extraction
func TestExtractHTMLSecurityHits(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		wantSubstrs []string
	}{
		{
			name: "extracts title, form, input, link, script and comment",
			html: `
				<html>
				<head><title>Admin Panel</title></head>
				<body>
					<!-- TODO: remove debug endpoint before release -->
					<form action="/login" method="post">
						<input name="username" type="text" value="">
					</form>
					<a href="https://external.example.com/track">link</a>
					<script src="/static/app.js"></script>
				</body>
				</html>`,
			wantSubstrs: []string{
				"Title: Admin Panel",
				"Method: POST | Action: /login",
				"[Input] Name: username | Type: text",
				"https://external.example.com/track",
				"/static/app.js",
				"remove debug endpoint",
			},
		},
		{
			name:        "empty document reports no findings",
			html:        `<html><body></body></html>`,
			wantSubstrs: []string{"No specific HTML security elements found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractHTMLSecurityHits([]byte(tt.html))

			for _, want := range tt.wantSubstrs {
				if !strings.Contains(got, want) {
					t.Errorf("ExtractHTMLSecurityHits() missing %q\nfull output:\n%s", want, got)
				}
			}
		})
	}
}

// TestFormatBody covers content-type-based formatting and the raw fallback
func TestFormatBody(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantSubstr  string
	}{
		{
			name:        "json is pretty printed",
			contentType: "application/json",
			body:        `{"a":1,"b":2}`,
			wantSubstr:  "\"a\": 1",
		},
		{
			name:        "form url encoded is parsed into key = value pairs",
			contentType: "application/x-www-form-urlencoded",
			body:        "user=admin&pass=hunter2",
			wantSubstr:  "user = [admin]",
		},
		{
			name:        "html falls through to security hit extraction",
			contentType: "text/html",
			body:        "<html><head><title>Test</title></head></html>",
			wantSubstr:  "Title: Test",
		},
		{
			name:        "unknown content type falls back to raw output",
			contentType: "application/octet-stream",
			body:        "raw-bytes-here",
			wantSubstr:  "raw-bytes-here",
		},
		{
			name:        "empty content type falls back to raw output",
			contentType: "",
			body:        "no-content-type",
			wantSubstr:  "no-content-type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBody(tt.contentType, []byte(tt.body))

			if !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("formatBody(%q, %q) = %q, want substring %q", tt.contentType, tt.body, got, tt.wantSubstr)
			}
		})
	}
}

// TestAuditRequest covers the top-level orchestration, including the nil-response
// guard and the nil-Request guard added during the refactor
func TestAuditRequest(t *testing.T) {
	t.Run("nil response is handled without panicking", func(t *testing.T) {
		var buf bytes.Buffer
		AuditRequest(&buf, nil)

		if !strings.Contains(buf.String(), "Cannot audit a nil response") {
			t.Errorf("AuditRequest(nil) output = %q, want nil-response message", buf.String())
		}
	})

	t.Run("response with nil Request does not panic and reports unknown host", func(t *testing.T) {
		resp := &http.Response{
			Status:     "200 OK",
			Proto:      "HTTP/1.1",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Request:    nil, // deliberately unset to exercise the nil guard
		}

		var buf bytes.Buffer
		AuditRequest(&buf, resp)
		got := buf.String()

		if !strings.Contains(got, "target: unknown") {
			t.Errorf("AuditRequest() output = %q, want target: unknown", got)
		}
	})

	t.Run("response body is scanned and formatted", func(t *testing.T) {
		resp := &http.Response{
			Status: "200 OK",
			Proto:  "HTTP/1.1",
			Header: http.Header{"Content-Type": []string{"application/json"}},
			Body:   io.NopCloser(strings.NewReader(`{"aws_secret_access_key":"AKIAABCDEFGHIJKLMNOP"}`)),
		}

		var buf bytes.Buffer
		AuditRequest(&buf, resp)
		got := buf.String()

		if !strings.Contains(got, "SENSITIVE DATA LEAKS DETECTED") {
			t.Errorf("AuditRequest() output = %q, want a sensitive data leak warning", got)
		}
		if !strings.Contains(got, "Body Analysis") {
			t.Errorf("AuditRequest() output = %q, want a Body Analysis section", got)
		}
	})

	t.Run("empty body skips body analysis section", func(t *testing.T) {
		resp := &http.Response{
			Status: "204 No Content",
			Proto:  "HTTP/1.1",
			Header: http.Header{},
			Body:   io.NopCloser(strings.NewReader("")),
		}

		var buf bytes.Buffer
		AuditRequest(&buf, resp)
		got := buf.String()

		if strings.Contains(got, "Body Analysis") {
			t.Errorf("AuditRequest() output = %q, expected no Body Analysis section for an empty body", got)
		}
	})
}