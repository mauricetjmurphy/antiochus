package server

import (
	"html"
	"net/http"
	"regexp"
	"strings"
)

// renderMarkdownToHTML converts simple markdown to styled HTML.
// Handles headers, paragraphs, code blocks, inline code, tables, bold, links, and lists.
func renderMarkdownToHTML(md string) string {
	lines := strings.Split(md, "\n")
	var out strings.Builder

	out.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Antiochus Setup Guide</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&family=Space+Grotesk:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { background: #0a0c10; color: #e5e7eb; font-family: 'Space Grotesk', system-ui, sans-serif; line-height: 1.7; padding: 2rem; }
  .container { max-width: 720px; margin: 0 auto; }
  h1 { color: #00ff88; font-size: 2rem; margin: 2rem 0 1rem; }
  h2 { color: #00ff88; font-size: 1.4rem; margin: 2rem 0 0.75rem; border-bottom: 1px solid #1e2028; padding-bottom: 0.5rem; }
  h3 { color: #00cc6a; font-size: 1.1rem; margin: 1.5rem 0 0.5rem; }
  p { margin: 0.75rem 0; }
  a { color: #00ff88; text-decoration: none; }
  a:hover { text-decoration: underline; }
  strong { color: #fff; }
  code { font-family: 'JetBrains Mono', monospace; background: #111318; padding: 0.15rem 0.4rem; border-radius: 4px; font-size: 0.85em; color: #00ff88; }
  pre { background: #111318; border: 1px solid #1e2028; border-radius: 8px; padding: 1rem; margin: 1rem 0; overflow-x: auto; }
  pre code { background: none; padding: 0; color: #e5e7eb; }
  ul, ol { margin: 0.75rem 0; padding-left: 1.5rem; }
  li { margin: 0.3rem 0; }
  table { width: 100%%; border-collapse: collapse; margin: 1rem 0; }
  th, td { text-align: left; padding: 0.5rem 0.75rem; border: 1px solid #1e2028; }
  th { background: #111318; color: #00ff88; font-size: 0.85rem; }
  td { font-size: 0.9rem; }
  hr { border: none; border-top: 1px solid #1e2028; margin: 2rem 0; }
  .back { display: inline-block; margin-bottom: 1.5rem; color: #00ff88; font-size: 0.9rem; }
</style>
</head>
<body>
<div class="container">
<a href="/" class="back">&larr; Back to Antiochus</a>
`)

	inCodeBlock := false
	inTable := false
	inList := false

	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	codeRe := regexp.MustCompile("`([^`]+)`")

	formatInline := func(s string) string {
		s = html.EscapeString(s)
		s = linkRe.ReplaceAllString(s, `<a href="$2">$1</a>`)
		s = boldRe.ReplaceAllString(s, `<strong>$1</strong>`)
		s = codeRe.ReplaceAllString(s, `<code>$1</code>`)
		return s
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code blocks
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				out.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				out.WriteString("<pre><code>")
				inCodeBlock = true
			}
			continue
		}
		if inCodeBlock {
			out.WriteString(html.EscapeString(line) + "\n")
			continue
		}

		// Close table if we're leaving it
		if inTable && !strings.HasPrefix(trimmed, "|") {
			out.WriteString("</table>\n")
			inTable = false
		}

		// Close list if we're leaving it
		if inList && !strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(trimmed, "* ") &&
			!(len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && trimmed[1] == '.') {
			out.WriteString("</ul>\n")
			inList = false
		}

		// Empty line
		if trimmed == "" {
			continue
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "***" {
			out.WriteString("<hr>\n")
			continue
		}

		// Headers
		if strings.HasPrefix(trimmed, "### ") {
			out.WriteString("<h3>" + formatInline(trimmed[4:]) + "</h3>\n")
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			out.WriteString("<h2>" + formatInline(trimmed[3:]) + "</h2>\n")
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			out.WriteString("<h1>" + formatInline(trimmed[2:]) + "</h1>\n")
			continue
		}

		// Table rows
		if strings.HasPrefix(trimmed, "|") {
			cells := strings.Split(trimmed, "|")
			// Skip separator rows like |---|---|
			if len(cells) > 2 && strings.TrimSpace(cells[1]) != "" &&
				strings.Trim(strings.TrimSpace(cells[1]), "-: ") != "" {
				if !inTable {
					out.WriteString("<table>\n")
					inTable = true
					// First row is header
					out.WriteString("<tr>")
					for _, c := range cells[1 : len(cells)-1] {
						out.WriteString("<th>" + formatInline(strings.TrimSpace(c)) + "</th>")
					}
					out.WriteString("</tr>\n")
				} else {
					out.WriteString("<tr>")
					for _, c := range cells[1 : len(cells)-1] {
						out.WriteString("<td>" + formatInline(strings.TrimSpace(c)) + "</td>")
					}
					out.WriteString("</tr>\n")
				}
			}
			continue
		}

		// Unordered list
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList {
				out.WriteString("<ul>\n")
				inList = true
			}
			out.WriteString("<li>" + formatInline(trimmed[2:]) + "</li>\n")
			continue
		}

		// Ordered list
		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && trimmed[1] == '.' {
			if !inList {
				out.WriteString("<ul>\n")
				inList = true
			}
			out.WriteString("<li>" + formatInline(strings.TrimSpace(trimmed[2:])) + "</li>\n")
			continue
		}

		// Paragraph
		out.WriteString("<p>" + formatInline(trimmed) + "</p>\n")
	}

	if inCodeBlock {
		out.WriteString("</code></pre>\n")
	}
	if inTable {
		out.WriteString("</table>\n")
	}
	if inList {
		out.WriteString("</ul>\n")
	}

	out.WriteString("</div>\n</body>\n</html>")
	return out.String()
}

func (s *Server) handleGuide(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(s.setupGuideHTML))
}
