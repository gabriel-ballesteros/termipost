package model

import (
	"strings"
	"testing"
)

const nestedJSON = `{"a":1,"b":{"c":2,"d":[1,2]},"e":{},"f":"x"}`

func seedFold(t *testing.T, s string) *foldView {
	t.Helper()
	f := &foldView{}
	f.setJSON(s)
	return f
}

func TestFoldDetectMultiLineOnly(t *testing.T) {
	f := seedFold(t, nestedJSON)
	if !f.foldable {
		t.Fatal("nested JSON should be foldable")
	}
	// Prettified:
	// 0 {
	// 1   "a": 1,
	// 2   "b": {
	// 3     "c": 2,
	// 4     "d": [
	// 5       1,
	// 6       2
	// 7     ],
	// 8     "e_placeholder?"  -> actually "e": {} stays single line
	// Verify the empty object "e": {} produced no region and the array/object did.
	if _, ok := f.headerOf[0]; !ok {
		t.Error("root object should be a region")
	}
	for h := range f.headerOf {
		if f.headerOf[h] <= h+1 {
			t.Errorf("region header %d -> close %d is not multi-line", h, f.headerOf[h])
		}
	}
	// "e": {} must not be a header.
	for i, ln := range f.lines {
		if strings.Contains(ln, `"e"`) {
			if _, ok := f.headerOf[i]; ok {
				t.Errorf("empty object line %d should not be foldable", i)
			}
		}
	}
}

func TestFoldNonJSON(t *testing.T) {
	for _, s := range []string{"", "   ", "hello world", "<html></html>", "{not json"} {
		f := seedFold(t, s)
		if f.foldable {
			t.Errorf("non-JSON %q should not be foldable", s)
		}
	}
}

func TestFoldVisibleLinesHidesCollapsedBody(t *testing.T) {
	f := seedFold(t, nestedJSON)
	full := len(f.visibleOrder())
	// Collapse the root: only the root header line remains visible.
	f.collapsed[0] = true
	vis := f.visibleOrder()
	if len(vis) != 1 || vis[0] != 0 {
		t.Fatalf("collapsed root should show only header, got %v", vis)
	}
	f.collapsed[0] = false
	if len(f.visibleOrder()) != full {
		t.Fatal("expanding root should restore all lines")
	}
}

func TestFoldNestedHiddenUnderCollapsedParent(t *testing.T) {
	f := seedFold(t, nestedJSON)
	// Find the "b" object header and its inner "d" array header.
	var bHeader, dHeader int = -1, -1
	for i, ln := range f.lines {
		if strings.Contains(ln, `"b"`) {
			if _, ok := f.headerOf[i]; ok {
				bHeader = i
			}
		}
		if strings.Contains(ln, `"d"`) {
			if _, ok := f.headerOf[i]; ok {
				dHeader = i
			}
		}
	}
	if bHeader < 0 || dHeader < 0 {
		t.Fatalf("expected b and d headers, got b=%d d=%d", bHeader, dHeader)
	}
	f.collapsed[bHeader] = true
	for _, li := range f.visibleOrder() {
		if li == dHeader {
			t.Fatal("nested d header should be hidden under collapsed b")
		}
	}
}

func TestFoldCursorSkipsHidden(t *testing.T) {
	f := seedFold(t, nestedJSON)
	f.collapsed[0] = true // collapse root, only line 0 visible
	f.cursor = 0
	f.moveDown() // nowhere else visible -> stays
	if f.cursor != 0 {
		t.Fatalf("cursor should stay on only visible line, got %d", f.cursor)
	}
	f.collapsed[0] = false
	f.cursor = 0
	f.moveDown()
	if f.cursor != 1 {
		t.Fatalf("cursor should move to line 1, got %d", f.cursor)
	}
}

func TestFoldToggleFromHeaderAndBody(t *testing.T) {
	f := seedFold(t, nestedJSON)
	// Toggle from header line 0.
	f.cursor = 0
	f.toggle()
	if !f.collapsed[0] {
		t.Fatal("toggle on header should collapse it")
	}
	f.toggle()
	if f.collapsed[0] {
		t.Fatal("second toggle should expand")
	}
	// Toggle from a body line collapses the nearest enclosing region.
	f.cursor = 1 // "a": 1 inside root
	f.toggle()
	if !f.collapsed[0] {
		t.Fatal("toggle from body should collapse enclosing region")
	}
	if f.cursor != 0 {
		t.Fatalf("cursor should park on header, got %d", f.cursor)
	}
}

func TestFoldRenderGutterMarkers(t *testing.T) {
	f := seedFold(t, nestedJSON)
	rows, _ := f.renderLines(false)
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
	// Root header expanded -> '-' marker present somewhere on the first row.
	if !strings.Contains(rows[0], "-") {
		t.Errorf("expanded header should show '-', got %q", rows[0])
	}
	// Collapse root and confirm '+' and ellipsis placeholder appear.
	f.collapsed[0] = true
	rows, _ = f.renderLines(false)
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "+") {
		t.Errorf("collapsed header should show '+', got %q", joined)
	}
	if !strings.Contains(joined, "…") {
		t.Errorf("collapsed header should show ellipsis, got %q", joined)
	}
}

func TestFoldCollapsedKeepsTrailingComma(t *testing.T) {
	// "b" object is followed by more members, so its close line ends with a comma.
	f := seedFold(t, nestedJSON)
	var bHeader int = -1
	for i, ln := range f.lines {
		if strings.Contains(ln, `"b"`) {
			if _, ok := f.headerOf[i]; ok {
				bHeader = i
			}
		}
	}
	if bHeader < 0 {
		t.Fatal("no b header")
	}
	f.cursor = bHeader
	f.collapsed[bHeader] = true
	line := f.lineContent(bHeader, false)
	if !strings.Contains(line, "},") {
		t.Errorf("collapsed b should retain trailing comma, got %q", line)
	}
}

func TestFoldRowRangeAccountsForWrap(t *testing.T) {
	f := seedFold(t, nestedJSON)
	// A narrow width forces lines to wrap; cursor on a later line should have a
	// row range beyond its plain line index.
	f.cursor = f.visibleOrder()[len(f.visibleOrder())-1]
	top, bottom := f.rowRange(4)
	if bottom < top {
		t.Fatalf("bottom %d < top %d", bottom, top)
	}
	if top == 0 {
		t.Error("last line at narrow width should have a non-zero top row")
	}
}

func TestFoldMoveUp(t *testing.T) {
	f := seedFold(t, nestedJSON)
	f.cursor = 2
	f.moveUp()
	if f.cursor != 1 {
		t.Fatalf("moveUp from 2 should land on 1, got %d", f.cursor)
	}
	f.cursor = 0
	f.moveUp() // already at top, stays
	if f.cursor != 0 {
		t.Fatalf("moveUp at top should stay at 0, got %d", f.cursor)
	}
}

func TestFoldMovesAndTogglesAreNoOpWhenNotFoldable(t *testing.T) {
	f := seedFold(t, "plain text body")
	if f.foldable {
		t.Fatal("non-JSON must not be foldable")
	}
	// None of these should panic or mutate state.
	f.moveUp()
	f.moveDown()
	f.toggle()
	if len(f.collapsed) != 0 || f.cursor != 0 {
		t.Fatalf("non-foldable view should stay inert (collapsed=%v cursor=%d)", f.collapsed, f.cursor)
	}
}

func TestFoldHiddenCursorSnapsBack(t *testing.T) {
	f := seedFold(t, nestedJSON)
	f.collapsed[0] = true // only line 0 visible
	f.cursor = 2          // pointing at a now-hidden line
	f.moveDown()          // cursorPos must snap to the only visible line
	if f.cursor != 0 {
		t.Fatalf("hidden cursor should snap to a visible line, got %d", f.cursor)
	}
}

func TestFoldRowRangeHiddenCursorAndZeroWidth(t *testing.T) {
	f := seedFold(t, nestedJSON)
	// Zero width: wrapHeight floors at 1 row.
	top, bottom := f.rowRange(0)
	if top != 0 || bottom != 0 {
		t.Fatalf("cursor on line 0 at width 0 should be rows (0,0), got (%d,%d)", top, bottom)
	}
	// Cursor hidden inside a collapsed section -> not found among visible lines.
	f.collapsed[0] = true
	f.cursor = 2
	if top, bottom := f.rowRange(80); top != 0 || bottom != 0 {
		t.Fatalf("hidden cursor row range should be (0,0), got (%d,%d)", top, bottom)
	}
}

func TestWindowRows(t *testing.T) {
	rows := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	// Fits: returned unchanged.
	if got := windowRows(rows, 3, 20); len(got) != len(rows) {
		t.Fatalf("rows fitting height should be unchanged, got %d", len(got))
	}
	// h <= 0: unchanged.
	if got := windowRows(rows, 3, 0); len(got) != len(rows) {
		t.Fatalf("non-positive height should return all rows, got %d", len(got))
	}
	cases := []struct {
		cursor    int
		wantFirst string
		wantLast  string
	}{
		{0, "0", "3"}, // clamp to start
		{5, "3", "6"}, // centered
		{9, "6", "9"}, // clamp to end
	}
	for _, c := range cases {
		got := windowRows(rows, c.cursor, 4)
		if len(got) != 4 {
			t.Fatalf("cursor %d: window len = %d, want 4", c.cursor, len(got))
		}
		if got[0] != c.wantFirst || got[len(got)-1] != c.wantLast {
			t.Errorf("cursor %d: window = %v, want first %q last %q", c.cursor, got, c.wantFirst, c.wantLast)
		}
	}
}

func TestFoldCacheResetsOnChange(t *testing.T) {
	f := seedFold(t, nestedJSON)
	f.collapsed[0] = true
	f.setJSON(nestedJSON) // unchanged -> cached, state preserved
	if !f.collapsed[0] {
		t.Error("unchanged source should preserve fold state")
	}
	f.setJSON(`{"x":{"y":1}}`) // changed -> reset
	if f.collapsed[0] {
		t.Error("changed source should reset collapsed state")
	}
	if f.cursor != 0 {
		t.Error("changed source should reset cursor")
	}
}
