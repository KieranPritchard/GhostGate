package commands

import (
	"GhostGate/internal/util"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// Finding represents a detected secret or sensitive data pattern
type Finding struct {
	Type  string // Type of secret detected (e.g. AWS Key, JWT)
	Match string // The matched string
}

// SecretRule defines a signature to match against response bodies
type SecretRule struct {
	Type    string
	Pattern *regexp.Regexp
}

// ---------------------------------------------------------------------------
// Reference Data
// ---------------------------------------------------------------------------

// DefaultRules includes signatures for high-risk tokens and credentials
var DefaultRules = []SecretRule{
	{
		Type:    "AWS Access Key ID",
		Pattern: regexp.MustCompile(`\b(AKIA|ASIA|ABIA|ACCA)[0-9A-Z]{16}\b`),
	},
	{
		Type:    "AWS Secret Access Key",
		Pattern: regexp.MustCompile(`(?i)aws_secret_access_key["'\s:=]+([a-zA-Z0-9/+=]{40})`),
	},
	{
		Type:    "JSON Web Token (JWT)",
		Pattern: regexp.MustCompile(`\beyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\b`),
	},
	{
		Type:    "Generic API Key Assignment",
		Pattern: regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|auth[_-]?token)["'\s:=]+["']([a-zA-Z0-9_\-]{16,64})["']`),
	},
	{
		Type:    "GitHub Personal Access Token",
		Pattern: regexp.MustCompile(`\b(ghp|gho|ghu|ghs|ghr)_[a-zA-Z0-9]{36}\b`),
	},
	{
		Type:    "Slack Token",
		Pattern: regexp.MustCompile(`xox[baprs]-[0-9a-zA-Z]{10,48}`),
	},
	{
		Type:    "Private Key",
		Pattern: regexp.MustCompile(`-----BEGIN (RSA|EC|DSA|OPENSSH) PRIVATE KEY-----`),
	},
	{
		Type:    "Basic / Bearer Auth Header Leak",
		Pattern: regexp.MustCompile(`(?i)Authorization:\s*(Basic|Bearer)\s+[a-zA-Z0-9._\-+/=]+`),
	},
}

// securityHeaderRule pairs a recommended security header with what it protects against
type securityHeaderRule struct {
	Name string
	Desc string
}

// securityHeaderOrder lists the recommended security headers in a fixed order.
// A slice is used instead of a map so the "missing headers" output is always
// printed in the same, predictable order across runs.
var securityHeaderOrder = []securityHeaderRule{
	{"Strict-Transport-Security", "Enforces HTTPS (HSTS)"},
	{"Content-Security-Policy", "Mitigates XSS and data injection"},
	{"X-Frame-Options", "Prevents Clickjacking (DENY or SAMEORIGIN)"},
	{"X-Content-Type-Options", "Prevents MIME-sniffing (nosniff)"},
	{"Referrer-Policy", "Controls referrer information leakage"},
	{"Permissions-Policy", "Restricts browser features/APIs"},
}

// ---------------------------------------------------------------------------
// Header Auditing
// ---------------------------------------------------------------------------

// checkInsecureHeader flags insecure configurations or values in existing headers
func checkInsecureHeader(key, value string) string {
	key = strings.ToLower(key)
	value = strings.ToLower(value)

	switch key {
	case "server", "x-powered-by", "x-aspnet-version", "x-aspnetmvc-version":
		return " [!] WARNING: Information disclosure (exposes tech stack/version)"

	case "access-control-allow-origin":
		if value == "*" {
			return " [!] WARNING: Wildcard CORS policy (allows any domain to read responses)"
		}

	case "set-cookie":
		var warnings []string
		if !strings.Contains(value, "secure") {
			warnings = append(warnings, "missing Secure flag")
		}
		if !strings.Contains(value, "httponly") {
			warnings = append(warnings, "missing HttpOnly flag")
		}
		if !strings.Contains(value, "samesite") {
			warnings = append(warnings, "missing SameSite attribute")
		}
		if len(warnings) > 0 {
			return fmt.Sprintf(" [!] WARNING: Insecure cookie (%s)", strings.Join(warnings, ", "))
		}

	case "x-xss-protection":
		if value == "0" {
			return " [!] WARNING: XSS auditor disabled"
		}
		return " [i] NOTE: Deprecated header; prefer Content-Security-Policy"

	case "content-security-policy":
		if strings.Contains(value, "unsafe-inline") || strings.Contains(value, "unsafe-eval") {
			return " [!] WARNING: Weak CSP policy allows unsafe execution ('unsafe-inline' / 'unsafe-eval')"
		}
	}

	return ""
}

// AuditHeaders prints existing headers, flags insecure values, and identifies
// missing security headers. Output is written to w rather than directly to
// stdout, so callers (and tests) can redirect it wherever they need to.
func AuditHeaders(w io.Writer, header http.Header) {
	// Outputs the headers
	fmt.Fprintln(w, "[+] Request Headers:")

	// Tracks which recommended security headers are present
	foundHeaders := make(map[string]bool)

	// Sorts header keys first, since Go's map iteration order is randomised
	// and we want stable, repeatable output between runs
	keys := make([]string, 0, len(header))
	for key := range header {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Loops over the sorted keys and their values
	for _, key := range keys {
		joinedVal := strings.Join(header[key], ", ")

		// Checks if the header matches a known security header
		for _, sec := range securityHeaderOrder {
			if strings.EqualFold(key, sec.Name) {
				foundHeaders[sec.Name] = true
				break
			}
		}

		// Flags insecure values or leaks
		warning := checkInsecureHeader(key, joinedVal)

		// Outputs key, value, and any warnings
		fmt.Fprintf(w, "    %s: %s%s\n", key, joinedVal, warning)
	}

	// Checks for missing security headers, in the fixed reference order
	var missing []string
	for _, sec := range securityHeaderOrder {
		if !foundHeaders[sec.Name] {
			missing = append(missing, fmt.Sprintf("    - [%s]: %s", sec.Name, sec.Desc))
		}
	}

	// Outputs missing security headers if any were found
	if len(missing) > 0 {
		fmt.Fprintln(w, "\n[!] MISSING SECURITY HEADERS:")
		for _, msg := range missing {
			fmt.Fprintln(w, msg)
		}
	}
}

// ---------------------------------------------------------------------------
// Secret Detection
// ---------------------------------------------------------------------------

// ScanResponseBody checks raw bytes against defined secret detection rules
func ScanResponseBody(bodyBytes []byte) []Finding {
	var findings []Finding

	// Loops through each secret rule
	for _, rule := range DefaultRules {
		// Finds all matches in the body
		matches := rule.Pattern.FindAll(bodyBytes, -1)
		for _, match := range matches {
			findings = append(findings, Finding{
				Type:  rule.Type,
				Match: string(match),
			})
		}
	}

	return findings
}

// FormatFindings formats detected secrets for output
func FormatFindings(findings []Finding) string {
	// Checks if there are no findings
	if len(findings) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("[!] SENSITIVE DATA LEAKS DETECTED:\n")

	// Loops over findings
	for _, f := range findings {
		val := f.Match
		// Truncates long findings so the terminal output stays readable
		if len(val) > 60 {
			val = val[:57] + "..."
		}
		b.WriteString(fmt.Sprintf("    - [%s]: %s\n", f.Type, val))
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// Orchestration
// ---------------------------------------------------------------------------

// AuditRequest audits a completed HTTP response: it reports headers, scans
// the body for secret leaks, and prints a formatted body summary. The
// response body is always closed once reading has finished. Output is
// written to w rather than directly to stdout.
func AuditRequest(w io.Writer, response *http.Response) {
	// Checks if there is not a response
	if response == nil {
		fmt.Fprintln(w, "[!] Cannot audit a nil response.")
		return
	}
	// Closes the response body when done
	defer response.Body.Close()

	// Creates the header for the command; guards against a nil Request/URL,
	// which can happen if the response wasn't obtained via http.Client
	host := "unknown"
	if response.Request != nil && response.Request.URL != nil {
		host = response.Request.URL.Host
	}
	fmt.Fprintf(w, "[+] HTTP(S) Server Audit for target: %s\n", host)
	fmt.Fprintf(w, "[+] Protocol: %s\n", response.Proto)
	fmt.Fprintf(w, "[+] Status: %s\n", response.Status)

	// Audits response headers and checks for missing/insecure entries
	AuditHeaders(w, response.Header)

	// Reads the body
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(w, "\n[!] Failed to read response body: %v\n", err)
		return
	}

	// Skips body analysis entirely for empty bodies
	if len(bodyBytes) == 0 {
		return
	}

	// Scans the response body for sensitive data leaks
	if findings := ScanResponseBody(bodyBytes); len(findings) > 0 {
		fmt.Fprint(w, FormatFindings(findings))
	}

	// Outputs a header
	fmt.Fprintln(w, "\n[+] Body Analysis:")

	// Gets the content type
	contentType := response.Header.Get("Content-Type")

	// Formats and prints body based on detected content type
	fmt.Fprintln(w, util.FormatBody(contentType, bodyBytes))
}
