package components

import (
	"encoding/xml"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/testmode"
	"github.com/preslavrachev/gomjml/parser"
)

func TestNavbarIDGeneration(t *testing.T) {
	// Lock test mode for exclusive access during this test
	ctrl := testmode.LockForTesting()
	defer ctrl.Release()

	t.Run("non-deterministic IDs when test mode disabled", func(t *testing.T) {
		ctrl.Disable()
		resetNavbarTestIndex()

		node := &parser.MJMLNode{XMLName: xml.Name{Local: "mj-navbar"}}
		opts := &options.RenderOpts{}

		seenIDs := make(map[string]bool)
		iterations := 10

		for i := range iterations {
			c := NewMJNavbarComponent(node, opts)
			id := c.generateCheckboxID()

			// ID should be 16 characters
			if len(id) != 16 {
				t.Errorf("Iteration %d: Expected 16-char hex string, got length: %d", i, len(id))
			}

			// ID should be unique (not seen before)
			if seenIDs[id] {
				t.Errorf("Iteration %d: Got duplicate ID: %s", i, id)
			}

			seenIDs[id] = true
		}

		// Verify we actually generated the expected number of unique IDs
		if len(seenIDs) != iterations {
			t.Errorf("Expected %d unique IDs, got %d", iterations, len(seenIDs))
		}
	})

	t.Run("deterministic IDs when test mode enabled", func(t *testing.T) {
		ctrl.Enable()
		resetNavbarTestIndex()

		// Create multiple navbar components
		node := &parser.MJMLNode{XMLName: xml.Name{Local: "mj-navbar"}}
		opts := &options.RenderOpts{}

		c1 := NewMJNavbarComponent(node, opts)
		c2 := NewMJNavbarComponent(node, opts)
		c3 := NewMJNavbarComponent(node, opts)

		id1 := c1.generateCheckboxID()
		id2 := c2.generateCheckboxID()
		id3 := c3.generateCheckboxID()

		// IDs should match the hardcoded test fixtures
		expectedIDs := navbarTestIDs
		if id1 != expectedIDs[0] {
			t.Errorf("Expected first ID to be %s, got %s", expectedIDs[0], id1)
		}
		if id2 != expectedIDs[1] {
			t.Errorf("Expected second ID to be %s, got %s", expectedIDs[1], id2)
		}
		if id3 != expectedIDs[2] {
			t.Errorf("Expected third ID to be %s, got %s", expectedIDs[2], id3)
		}
	})

	t.Run("falls back to random after exhausting test IDs", func(t *testing.T) {
		ctrl.Enable()
		resetNavbarTestIndex()

		node := &parser.MJMLNode{XMLName: xml.Name{Local: "mj-navbar"}}
		opts := &options.RenderOpts{}

		// Exhaust all test IDs
		for i := 0; i < len(navbarTestIDs); i++ {
			c := NewMJNavbarComponent(node, opts)
			c.generateCheckboxID()
		}

		// Next ID should be random (16 chars but not from test IDs)
		c := NewMJNavbarComponent(node, opts)
		id := c.generateCheckboxID()

		if len(id) != 16 {
			t.Errorf("Expected 16-char hex string after exhausting test IDs, got length: %d", len(id))
		}

		// Verify it's not one of the test IDs
		for _, testID := range navbarTestIDs {
			if id == testID {
				t.Errorf("Expected random ID after exhausting test IDs, got test ID: %s", id)
			}
		}
	})
}
