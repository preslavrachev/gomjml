package mjml

import (
	"os"
	"regexp"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/components"
)

var textContentPattern = regexp.MustCompile(`(?s)<div style="font-family[^>]*>(.*?)</div></td>`)

// The integration suite compares DOM trees, which cannot see whitespace, so
// the content of each mj-text is compared with the golden byte for byte.
func TestMJTextKeepsPreformattedWhitespace(t *testing.T) {
	components.EnableTestMode()
	source, err := os.ReadFile("testdata/mj-text-pre.mjml")
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/mj-text-pre.html")
	if err != nil {
		t.Fatal(err)
	}
	actual, err := Render(string(source))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	want := textContentPattern.FindAllStringSubmatch(string(golden), -1)
	got := textContentPattern.FindAllStringSubmatch(actual, -1)
	if len(got) != len(want) {
		t.Fatalf("rendered %d mj-text blocks, golden has %d", len(got), len(want))
	}
	for i := range want {
		if got[i][1] != want[i][1] {
			t.Errorf("mj-text %d:\n got %q\nwant %q", i+1, got[i][1], want[i][1])
		}
	}
}
