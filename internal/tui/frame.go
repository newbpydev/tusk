package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/terminaltext"
	"github.com/rivo/uniseg"
)

// cellSlice keeps complete graphemes. If a clipping boundary intersects a wide
// cluster, its visible portion becomes spaces instead of half a glyph. Input
// must contain only sanitized text and trusted style sequences.
func cellSlice(s string, left, right int) string {
	if right <= left {
		return ""
	}
	left = max(0, left)
	var out strings.Builder
	pos := 0
	styled := false
	for len(s) > 0 {
		if s[0] == '\x1b' {
			seq, _, n, _ := ansi.DecodeSequence(s, 0, nil)
			out.WriteString(seq)
			styled = true
			s = s[n:]
			continue
		}
		seq, rest, w, _ := uniseg.FirstGraphemeClusterInString(s, -1)
		s = rest
		if w == 0 {
			if pos >= left && pos < right {
				out.WriteString(seq)
			}
			continue
		}
		end := pos + w
		if end > left && pos < right {
			if pos >= left && end <= right {
				out.WriteString(seq)
			} else {
				out.WriteString(strings.Repeat(" ", min(end, right)-max(pos, left)))
			}
		}
		pos = end
	}
	if styled {
		out.WriteString("\x1b[0m")
	}
	return out.String()
}

func fitCells(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = cellSlice(s, 0, width)
	return s + strings.Repeat(" ", max(0, width-ansi.StringWidth(s)))
}

func titleCells(s string, width int) string {
	if ansi.StringWidth(s) > width && width > 0 {
		return fitCells(cellSlice(s, 0, width-1)+"…", width)
	}
	return fitCells(s, width)
}

func (m *Model) panel(title string, lines []string, w, h int, focus bool) []string {
	if w < 2 || h < 2 {
		return nil
	}
	color := borderColor
	titleColor := mutedColor
	if focus {
		color = accentColor
		titleColor = accentColor
		title = "> " + title
	}
	edge := func(s string) string { return m.paint(s, color, false) }
	label := cellSlice(" "+title+" ", 0, max(0, w-3))
	top := edge("╭─") + m.paint(label, titleColor, true) + edge(strings.Repeat("─", max(0, w-3-ansi.StringWidth(label)))+"╮")
	rows := make([]string, 0, h)
	rows = append(rows, fitCells(top, w))
	for y := 0; y < h-2; y++ {
		s := ""
		if y < len(lines) {
			s = lines[y]
		}
		rows = append(rows, edge("│")+fitCells(s, w-2)+edge("│"))
	}
	return append(rows, edge("╰"+strings.Repeat("─", w-2)+"╯"))
}

func (m *Model) listLines(w, h int) []string {
	query := m.filter.SearchTerm
	if m.searching {
		query = m.searchDraft
	}
	search := m.paint(" /  Search tasks", mutedColor, false)
	if query != "" || m.searching {
		search = m.paint(" /  ", accentColor, true) + inputLine(query, m.searching, max(1, w-6))
	}
	all := "1 All tasks"
	if w < 40 {
		all = "1 All"
	}
	var tabs []string
	for i, label := range []string{all, "2 Today", "3 Done"} {
		if i == m.workspaceTab {
			label = m.paint("["+label+"]", accentColor, true)
		} else {
			label = m.paint(label, mutedColor, false)
		}
		tabs = append(tabs, label)
	}
	toolbar := []string{fitCells(search, w), fitCells(" "+strings.Join(tabs, "  "), w), m.paint(strings.Repeat("─", w), borderColor, false)}
	h = max(0, h-len(toolbar))
	withToolbar := func(lines []string) []string { return append(toolbar, lines...) }
	switch m.state {
	case loading:
		return withToolbar([]string{"", "  Loading tasks…"})
	case loadFailed:
		return withToolbar([]string{"", "  Could not load tasks.", "  Press r to retry."})
	}
	if len(m.forest) == 0 {
		return withToolbar([]string{"", "  No tasks", "  Your workspace is ready.", "", "  a  Create your first task"})
	}
	if len(m.rows) == 0 {
		return withToolbar([]string{"", "  No matching tasks", "  Esc clears filters."})
	}
	type entry struct {
		row, part int
		text      string
	}
	var entries []entry
	group := ""
	selectedLine := 0
	for i, row := range m.rows {
		if row.group != group {
			if group != "" {
				entries = append(entries, entry{row: -1})
			}
			group = row.group
			entries = append(entries, entry{row: -1, text: "  " + group})
		}
		if i == m.selected {
			selectedLine = len(entries)
		}
		entries = append(entries, entry{row: i}, entry{row: i, part: 1})
	}
	if selectedLine < m.listOffset {
		m.listOffset = selectedLine
	}
	if selectedLine+2 > m.listOffset+h {
		m.listOffset = selectedLine + 2 - h
	}
	m.listOffset = max(0, min(m.listOffset, max(0, len(entries)-h)))
	var lines []string
	for _, entry := range entries[m.listOffset:min(len(entries), m.listOffset+max(0, h))] {
		if entry.row < 0 {
			lines = append(lines, m.paint(fitCells(entry.text, w), mutedColor, true))
			continue
		}
		row := m.rows[entry.row]
		task := row.node.Task
		s := ""
		if entry.part == 1 {
			prefix := row.continuation
			if len(row.node.Children) > 0 {
				if row.depth == 0 {
					prefix += "  "
				}
				if entry.row+1 < len(m.rows) && m.rows[entry.row+1].depth > row.depth {
					prefix += "│ "
				} else {
					prefix += "  "
				}
			}
			if row.depth > 0 || len(row.node.Children) == 0 {
				prefix += "  "
			}
			_, status := statusText(task.Status)
			s = "  " + clippedTreeGuide(prefix, max(0, w-20)) + status + " · " + priorityLabel(task.Priority)
			if task.Progress > 0 {
				s += fmt.Sprintf(" · %d%%", task.Progress)
			}
		} else {
			context := ""
			if row.context {
				context = "[context] "
			}
			prefix := "  " + clippedTreeGuide(row.branch, max(2, w-12-len(context)))
			if len(row.node.Children) > 0 {
				if m.collapsed[task.ID] && !m.filter.HasPredicates() && m.dueStart == nil {
					prefix += "▸ "
				} else {
					prefix += "▾ "
				}
			}
			if entry.row == m.selected {
				prefix = " >" + prefix[2:]
			}
			mark, _ := statusText(task.Status)
			s = prefix + context + mark + " " + terminaltext.Scalar(task.Title)
		}
		s = titleCells(s, w)
		if entry.row == m.selected {
			fg := textColor
			if entry.part == 1 {
				fg = accentColor
			}
			s = m.surface(m.paint(s, fg, entry.part == 0), fg, selectionColor)
		} else if entry.part == 1 {
			s = m.paint(s, mutedColor, false)
		}
		lines = append(lines, s)
	}
	return withToolbar(lines)
}

// Deep paths yield space to the task's title and metadata in narrow panels.
func clippedTreeGuide(guide string, width int) string {
	if ansi.StringWidth(guide) <= width {
		return guide
	}
	return cellSlice(guide, 0, max(0, width-2)) + cellSlice("… ", 0, width)
}

// prepareFrame renders the current state. It never re-projects the forest:
// callers own rebuilds (finish rebuilds once per Update, New before the first
// frame), so a render can never duplicate or reorder projections.
func (m *Model) prepareFrame() {
	l := measure(m.width, m.height)
	if l.width == 0 || l.height == 0 {
		m.frame = ""
		return
	}
	if !l.usable() {
		lines := make([]string, l.height)
		for i := range lines {
			s := ""
			if i == 0 {
				s = "Resize to 80×24; Ctrl+C quits"
			}
			lines[i] = m.surface(fitCells(s, l.width), textColor, canvasColor)
		}
		m.frame = strings.Join(lines, "\n")
		return
	}
	if m.form != nil && m.form.calendar != nil {
		l.modal.height = min(l.modal.height, 18)
		l.modal.y = (l.height - l.modal.height) / 2
	}
	noun := "tasks"
	if len(m.rows) == 1 {
		noun = "task"
	}
	left := m.panel(fmt.Sprintf("Tasks · %d %s", len(m.rows), noun), m.listLines(l.listWidth-2, l.bodyHeight-2), l.listWidth, l.bodyHeight, m.focus == listFocus)
	details := m.detailLines(l.detailsWidth - 2)
	if m.detailsEnd {
		m.detailsScroll = max(0, len(details)-(l.bodyHeight-2))
	}
	m.detailsScroll = max(0, min(m.detailsScroll, max(0, len(details)-(l.bodyHeight-2))))
	detailTitle := "Task details"
	if len(details) > l.bodyHeight-2 {
		detailTitle += fmt.Sprintf(" · %d–%d / %d", m.detailsScroll+1, min(len(details), m.detailsScroll+l.bodyHeight-2), len(details))
	}
	right := m.panel(detailTitle, details[m.detailsScroll:], l.detailsWidth, l.bodyHeight, m.focus == detailsFocus)
	header := m.paint(" TUSK", accentColor, true) + m.paint("  /  Personal workspace", mutedColor, false)
	actions := m.paint("f Filters", mutedColor, false) + "   " + m.paint("a New task", accentColor, true) + " "
	if m.workspaceTab == -1 {
		label := "f Filters active"
		if m.dueLabel != "" {
			label = "f Filters · " + m.dueLabel
		}
		actions = m.paint(label, accentColor, true) + "   " + m.paint("a New task", accentColor, true) + " "
	}
	header = fitCells(header, max(0, l.width-ansi.StringWidth(actions))) + actions
	rows := []string{fitCells(header, l.width)}
	for y := 0; y < l.bodyHeight; y++ {
		rows = append(rows, left[y]+right[y])
	}
	for y, s := range rows {
		fg := textColor
		if m.helpOpen || m.filters != nil || m.form != nil || m.confirmation != nil || m.recoveryNeeded {
			s = ansi.Strip(s)
			fg = borderColor
		}
		rows[y] = m.surface(s, fg, canvasColor)
	}
	m.help.Width = l.width - 2
	footer := " " + m.help.ShortHelpView(browseHints())
	if m.notice != "" {
		hints := " · r refresh · ? help · q quit"
		footer = " " + strings.TrimRight(titleCells(m.notice, max(1, l.width-1-ansi.StringWidth(hints))), " ") + hints
	}
	if m.focus == detailsFocus {
		footer = " ↑↓ scroll · PgUp/PgDn · Home/End · Tab tasks · ? help · q quit"
	}
	if m.recoveryNeeded {
		footer = " r reload · Tab / Enter choose · ↑↓ PgUp/PgDn scroll · q quit"
		if m.recovery != nil && m.recovery.blocked {
			footer = " Recovery blocked · Enter / q quit · ↑↓ PgUp/PgDn scroll"
		}
	} else if m.saving {
		footer = " Saving… · Ctrl+C exits safely"
	} else if m.stale {
		footer = " Stale snapshot · writes paused · r refresh · ? help · q quit"
		if m.awaitingRead {
			footer = " " + m.writeNotice() + "; refresh failed · writes paused · r retry · q quit"
		}
	} else if m.busy && m.state == loaded {
		footer = " Refreshing ·" + footer
	}
	if m.state == loadFailed {
		footer = " r retry  ·" + footer
	}
	if m.helpOpen {
		footer = " ↑↓ scroll help  ·  Esc / ? / q close  ·  Ctrl+C exit"
	}
	if m.searching {
		footer = " Type to search  ·  Enter accept  ·  Esc restore"
	}
	if m.filters != nil {
		footer = " Tab next · ←→ choose · Space toggle · Ctrl+S apply · Esc cancel"
	}
	if m.form != nil && !m.recoveryNeeded {
		footer = " Tab / Shift+Tab fields · Ctrl+S save · Esc cancel · Ctrl+C exit"
		if m.form.calendar != nil {
			footer = " ←→ day · ↑↓ week · PgUp/PgDn month · t today · Enter choose · Esc back"
		}
	}
	if m.confirmation != nil && !m.recoveryNeeded {
		footer = " Tab choose · Space subtree · Enter confirm · Esc cancel · ↑↓ scroll"
	}
	if m.stale && !strings.Contains(footer, "Stale") {
		footer = " Stale ·" + footer
	}
	rows = append(rows, m.surface(fitCells(footer, l.width), textColor, footerColor))
	if m.helpOpen {
		content := helpLines()
		area := max(0, l.modal.height-6)
		m.helpScroll = min(max(0, m.helpScroll), max(0, len(content)-area))
		rows = m.overlay(rows, l, "Keyboard shortcuts", content, m.helpScroll, "Esc / ? / q close  ·  Ctrl+C exit")
	}
	if m.filters != nil {
		controls := m.actionButton("Apply filters", m.filters.field == 4, true, false) + "  " + m.actionButton("Clear all", m.filters.field == 5, false, false)
		rows = m.overlay(rows, l, "Filter tasks", m.filterLines(l.modal.width-6), 0, controls)
	}
	if m.form != nil && !m.recoveryNeeded {
		title, content, offset, controls := m.formContent(l.modal.width-6, l.modal.height-6)
		rows = m.overlay(rows, l, title, content, offset, controls)
		b := l.modal
		y := b.y + b.height - 4
		left := cellSlice(rows[y], 0, b.x+1)
		right := cellSlice(rows[y], b.x+b.width-1, l.width)
		rows[y] = left + m.surface("  "+fitCells(m.form.err, b.width-6)+"  ", warningColor, dialogColor) + right
	}
	if m.confirmation != nil && !m.recoveryNeeded {
		content, controls := m.confirmContent(l.modal.width - 6)
		m.confirmation.scroll = max(0, min(m.confirmation.scroll, max(0, len(content)-(l.modal.height-6))))
		rows = m.overlay(rows, l, "Delete task", content, m.confirmation.scroll, controls)
	}
	if m.recoveryNeeded && m.recovery != nil {
		content, offset, controls := m.recoveryContent(l.modal.width-6, l.modal.height-6)
		rows = m.overlay(rows, l, "Readback and recovery", content, offset, controls)
	}
	m.frame = strings.Join(rows, "\n")
}

// overlay keeps two horizontal cells and one vertical row of padding, with
// separate error and action rows so scrolling never hides the controls.
func (m *Model) overlay(rows []string, l layout, title string, content []string, offset int, controls string) []string {
	b := l.modal
	area := max(0, b.height-6)
	lines := make([]string, b.height-2)
	for i := 0; i < area; i++ {
		j := max(0, offset) + i
		if j < len(content) {
			lines[i+1] = "  " + fitCells(content[j], b.width-6) + "  "
		}
	}
	lines[len(lines)-2] = "  " + fitCells(controls, b.width-6) + "  "
	box := m.panel(title, lines, b.width, b.height, true)
	for i, row := range box {
		y := b.y + i
		left := fitCells(cellSlice(rows[y], 0, b.x), b.x)
		right := fitCells(cellSlice(rows[y], b.x+b.width, l.width), l.width-b.x-b.width)
		rows[y] = left + m.surface(row, textColor, dialogColor) + right
	}
	return rows
}
