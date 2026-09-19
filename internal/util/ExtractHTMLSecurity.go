package util

import (
	"bytes"
	"fmt"
	"strings"

	nethtml "golang.org/x/net/html"
)

// SecurityAnalysis contains key elements required during web assessments
type SecurityAnalysis struct {
	Title    string
	Meta     []string
	Forms    []string
	Links    []string
	Scripts  []string
	Comments []string
}

// ExtractHTMLSecurityHits parses HTML and extracts only key security indicators
func ExtractHTMLSecurityHits(htmlBytes []byte) string {
	doc, err := nethtml.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return "[!] Failed to parse HTML DOM"
	}

	analysis := &SecurityAnalysis{}

	var traverse func(*nethtml.Node)
	traverse = func(n *nethtml.Node) {
		// Extracts HTML comments (often leak endpoints, internal keys, or developer notes)
		if n.Type == nethtml.CommentNode {
			if comment := strings.TrimSpace(n.Data); comment != "" {
				analysis.Comments = append(analysis.Comments, comment)
			}
		}

		if n.Type == nethtml.ElementNode {
			switch strings.ToLower(n.Data) {
			case "title":
				if n.FirstChild != nil {
					analysis.Title = n.FirstChild.Data
				}

			// Extracts form method and action
			case "form":
				var action, method string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "action":
						action = attr.Val
					case "method":
						method = attr.Val
					}
				}
				analysis.Forms = append(analysis.Forms, fmt.Sprintf("Form -> Method: %s | Action: %s", strings.ToUpper(method), action))

			// Extracts input fields nested inside forms
			case "input":
				var name, typ, value string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "name":
						name = attr.Val
					case "type":
						typ = attr.Val
					case "value":
						value = attr.Val
					}
				}
				analysis.Forms = append(analysis.Forms, fmt.Sprintf("  [Input] Name: %s | Type: %s | Value: %s", name, typ, value))

			// Extracts anchor hrefs, useful for spotting external links or leaked internal paths
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						analysis.Links = append(analysis.Links, attr.Val)
					}
				}

			// Extracts external script sources
			case "script":
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						analysis.Scripts = append(analysis.Scripts, attr.Val)
					}
				}

			// Extracts meta name/property + content pairs
			case "meta":
				var name, content string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "name", "property":
						name = attr.Val
					case "content":
						content = attr.Val
					}
				}
				if name != "" {
					analysis.Meta = append(analysis.Meta, fmt.Sprintf("%s = %s", name, content))
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	var b strings.Builder
	if analysis.Title != "" {
		b.WriteString(fmt.Sprintf("Title: %s\n", analysis.Title))
	}

	if len(analysis.Forms) > 0 {
		b.WriteString("\n[+] Forms & Input Fields:\n")
		for _, f := range analysis.Forms {
			b.WriteString(fmt.Sprintf("    %s\n", f))
		}
	}

	if len(analysis.Links) > 0 {
		b.WriteString("\n[+] Links:\n")
		for _, l := range analysis.Links {
			b.WriteString(fmt.Sprintf("    %s\n", l))
		}
	}

	if len(analysis.Comments) > 0 {
		b.WriteString("\n[!] HTML Comments:\n")
		for _, c := range analysis.Comments {
			b.WriteString(fmt.Sprintf("    <!-- %s -->\n", c))
		}
	}

	if len(analysis.Scripts) > 0 {
		b.WriteString("\n[+] Script Sources:\n")
		for _, s := range analysis.Scripts {
			b.WriteString(fmt.Sprintf("    %s\n", s))
		}
	}

	if res := strings.TrimSpace(b.String()); res != "" {
		return res
	}
	return "[+] No specific HTML security elements found."
}
