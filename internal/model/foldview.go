package model

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/gabriel-ballesteros/termipost/internal/syntax"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// foldRegion is a multi-line JSON object/array spanning prettified lines
// [header, close]. Single-line and empty containers never form a region.
type foldRegion struct {
	header int
	close  int
}

// foldView renders a read-only JSON body as a foldable view: every multi-line
// object/array is an independently collapsible section marked with a left gutter
// (`-` expanded, `+` collapsed, blank otherwise). It tracks a line cursor over
// the visible lines and a per-section collapsed state. It is view-only and never
// mutates the source body. State resets whenever the source body changes.
type foldView struct {
	src      string
	lines    []string    // prettified body, one element per line (plain)
	colored  []string    // syntax-highlighted lines, aligned 1:1 with lines
	headerOf map[int]int  // header line -> matching close line (foldable regions only)
	regions  []foldRegion // all foldable regions, for enclosing lookups
	collapsed map[int]bool // header line -> collapsed?
	cursor   int          // line index of the cursor (always a visible line)
	foldable bool         // true when the body is JSON with >=1 foldable section
}

// setJSON re-seeds the view from s. It returns true when s is valid JSON with at
// least one foldable (multi-line) section; otherwise the view is non-foldable and
// callers should fall back to their plain display. Results are cached by source,
// so re-calling with an unchanged body is cheap and preserves fold state.
func (f *foldView) setJSON(s string) bool {
	if f.lines != nil && f.src == s {
		return f.foldable
	}
	*f = foldView{src: s, collapsed: map[int]bool{}, headerOf: map[int]int{}}
	if strings.TrimSpace(s) == "" || !looksLikeJSON([]byte(s)) {
		return false
	}
	pretty, err := syntax.Prettify(s)
	if err != nil {
		return false
	}
	f.lines = strings.Split(pretty, "\n")
	f.colored = strings.Split(syntax.HighlightJSON(pretty), "\n")
	if len(f.colored) != len(f.lines) {
		f.colored = f.lines // alignment safety net; degrade to plain
	}
	f.detect()
	f.foldable = len(f.regions) > 0
	return f.foldable
}

// detect finds every multi-line object/array via a single bracket-depth pass.
// A line whose last non-space char opens a container pushes its index; a line
// whose first non-space char closes one pops, forming a region kept only when it
// spans more than one line (excludes empty `{}`/`[]` and single-line containers).
func (f *foldView) detect() {
	var stack []int
	for i, ln := range f.lines {
		t := strings.TrimSpace(ln)
		if t == "" {
			continue
		}
		if c := t[0]; c == '}' || c == ']' {
			if n := len(stack); n > 0 {
				h := stack[n-1]
				stack = stack[:n-1]
				if i > h+1 {
					f.regions = append(f.regions, foldRegion{header: h, close: i})
					f.headerOf[h] = i
				}
			}
		}
		if c := t[len(t)-1]; c == '{' || c == '[' {
			stack = append(stack, i)
		}
	}
}

// visibleOrder returns the line indices that are currently shown, in order,
// skipping the body of every collapsed section.
func (f *foldView) visibleOrder() []int {
	out := make([]int, 0, len(f.lines))
	for i := 0; i < len(f.lines); {
		out = append(out, i)
		if cl, ok := f.headerOf[i]; ok && f.collapsed[i] {
			i = cl + 1
		} else {
			i++
		}
	}
	return out
}

func (f *foldView) moveUp()   { f.moveCursor(-1) }
func (f *foldView) moveDown() { f.moveCursor(1) }

func (f *foldView) moveCursor(d int) {
	vis := f.visibleOrder()
	if len(vis) == 0 {
		return
	}
	pos := f.cursorPos(vis) + d
	pos = max(0, min(pos, len(vis)-1))
	f.cursor = vis[pos]
}

// cursorPos returns the index of the cursor within the visible order, snapping to
// the nearest following visible line if the cursor became hidden.
func (f *foldView) cursorPos(vis []int) int {
	for i, l := range vis {
		if l >= f.cursor {
			return i
		}
	}
	return len(vis) - 1
}

// toggle collapses or expands the section at the cursor: the region whose header
// is the cursor line, else the nearest enclosing region. The cursor is parked on
// the header so it stays visible after collapsing.
func (f *foldView) toggle() {
	h, ok := f.regionFor(f.cursor)
	if !ok {
		return
	}
	f.collapsed[h] = !f.collapsed[h]
	f.cursor = h
}

// regionFor maps a line to the header of the region it controls: the line itself
// if it is a header, otherwise the smallest region that encloses it.
func (f *foldView) regionFor(line int) (int, bool) {
	if _, ok := f.headerOf[line]; ok {
		return line, true
	}
	best, bestSize := -1, int(^uint(0)>>1)
	for _, r := range f.regions {
		if line > r.header && line <= r.close {
			if sz := r.close - r.header; sz < bestSize {
				best, bestSize = r.header, sz
			}
		}
	}
	if best < 0 {
		return 0, false
	}
	return best, true
}

// renderLines returns the gutter-marked, highlighted visible lines in order plus
// the index (within those lines) of the cursor line. focused enables the cursor
// caret styling.
func (f *foldView) renderLines(focused bool) ([]string, int) {
	vis := f.visibleOrder()
	rows := make([]string, 0, len(vis))
	cursorRow := 0
	for idx, li := range vis {
		if li == f.cursor {
			cursorRow = idx
		}
		rows = append(rows, f.renderLine(li, focused && li == f.cursor))
	}
	return rows, cursorRow
}

func (f *foldView) renderLine(li int, isCursor bool) string {
	marker := " "
	if _, ok := f.headerOf[li]; ok {
		if f.collapsed[li] {
			marker = "+"
		} else {
			marker = "-"
		}
	}
	caret := " "
	if isCursor {
		caret = "▸"
	}
	gutter := caret + marker
	if isCursor {
		gutter = ui.FieldFocused.Render(gutter)
	} else {
		gutter = ui.Subtle.Render(gutter)
	}
	return gutter + " " + f.lineContent(li, true)
}

// lineContent returns the body of a display line. Collapsed section headers are
// rendered as `<header> … <close>` so any trailing comma on the close line is
// preserved. colored selects highlighted vs plain text.
func (f *foldView) lineContent(li int, colored bool) string {
	src := f.lines
	if colored {
		src = f.colored
	}
	if cl, ok := f.headerOf[li]; ok && f.collapsed[li] {
		indent := f.lines[li][:len(f.lines[li])-len(strings.TrimLeft(f.lines[li], " "))]
		ellipsis := " … "
		if colored {
			ellipsis = ui.Subtle.Render(ellipsis)
		}
		return indent + strings.TrimSpace(src[li]) + ellipsis + strings.TrimSpace(src[cl])
	}
	return src[li]
}

// rowRange returns the inclusive screen-row span of the cursor line within the
// rendered content at the given wrap width, accounting for soft-wrapped lines.
func (f *foldView) rowRange(width int) (int, int) {
	row := 0
	for _, li := range f.visibleOrder() {
		h := f.wrapHeight(li, width)
		if li == f.cursor {
			return row, row + h - 1
		}
		row += h
	}
	return 0, 0
}

func (f *foldView) wrapHeight(li, width int) int {
	if width <= 0 {
		return 1
	}
	plain := "   " + f.lineContent(li, false) // gutter (caret+marker+space) = 3 cols
	h := int(math.Ceil(float64(lipgloss.Width(plain)) / float64(width)))
	return max(h, 1)
}

// windowRows clamps rows to at most h lines, keeping the cursor row visible by
// centering the window around it. Used by panes that lack a scrollable viewport.
func windowRows(rows []string, cursor, h int) []string {
	if h <= 0 || len(rows) <= h {
		return rows
	}
	start := cursor - h/2
	start = max(0, min(start, len(rows)-h))
	return rows[start : start+h]
}
