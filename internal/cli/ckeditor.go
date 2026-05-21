package cli

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var ckeditorMarkdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

var ckeditorHTMLBlockPattern = regexp.MustCompile(`(?is)^\s*<(address|article|aside|blockquote|div|dl|fieldset|figcaption|figure|footer|form|h[1-6]|header|hr|main|nav|ol|p|pre|section|table|ul)(\s|>|/)`)

func formatCKEditorHTML(value string) string {
	trimmed := strings.TrimSpace(normalizeNewlines(value))
	if trimmed == "" {
		return ""
	}
	if looksLikeCKEditorHTML(trimmed) {
		return trimmed
	}

	var out bytes.Buffer
	if err := ckeditorMarkdown.Convert([]byte(trimmed), &out); err != nil {
		return trimmed
	}
	return strings.TrimSpace(out.String())
}

func normalizeNewlines(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.ReplaceAll(value, "\r", "\n")
}

func looksLikeCKEditorHTML(value string) bool {
	return ckeditorHTMLBlockPattern.MatchString(value)
}
