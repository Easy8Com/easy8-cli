package cli

import (
	"strings"
	"testing"
)

func TestFormatCKEditorHTMLPlainText(t *testing.T) {
	got := formatCKEditorHTML("Hotovo")
	if got != "<p>Hotovo</p>" {
		t.Fatalf("html = %q", got)
	}
}

func TestFormatCKEditorHTMLMarkdownList(t *testing.T) {
	got := formatCKEditorHTML("- Prvni\n- Druhy")
	if !strings.Contains(got, "<ul>") || !strings.Contains(got, "<li>Prvni</li>") || !strings.Contains(got, "<li>Druhy</li>") {
		t.Fatalf("html = %q", got)
	}
}

func TestFormatCKEditorHTMLFencedCode(t *testing.T) {
	got := formatCKEditorHTML("```bash\ngo test ./...\n```")
	if !strings.Contains(got, `<pre><code class="language-bash">go test ./...`) {
		t.Fatalf("html = %q", got)
	}
}

func TestFormatCKEditorHTMLPreservesBlockHTML(t *testing.T) {
	input := "<p><strong>Hotovo</strong></p>"
	got := formatCKEditorHTML(input)
	if got != input {
		t.Fatalf("html = %q", got)
	}
}
