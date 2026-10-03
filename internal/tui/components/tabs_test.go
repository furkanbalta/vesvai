package components

import "testing"

func TestTabsScrollKeepsActiveVisible(t *testing.T) {
	names := []string{"General", "Session", "MCP", "Skills", "Rules", "Plugins", "Permissions", "Memory", "System"}
	tabs := NewTabs(names)

	width := 40
	tabs.SetActive(8)
	start, end := tabs.fitWindow(width)
	if start >= 8 || end < 9 {
		t.Fatalf("window [%d,%d) does not contain active tab 8", start, end)
	}
	if tabs.scroll != start {
		t.Fatalf("scroll = %d, want %d", tabs.scroll, start)
	}

	tabs.SetActive(0)
	start, end = tabs.fitWindow(width)
	if start != 0 || end <= 0 {
		t.Fatalf("window [%d,%d) does not contain active tab 0", start, end)
	}

	tabs.SetActive(8)
	start, end = tabs.fitWindow(200)
	if start != 0 || end != len(names) {
		t.Fatalf("window [%d,%d), want [0,%d)", start, end, len(names))
	}
	if tabs.scroll != 0 {
		t.Fatalf("scroll = %d, want 0", tabs.scroll)
	}
}
