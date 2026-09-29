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
	if focus {
		color = accentColor
		title = "> " + title
	}
	edge := func(s string) string { return m.paint(s, color, false) }
	label := cellSlice(" "+title+" ", 0, max(0, w-3))
	top := edge("╭─") + m.paint(label, color, true) + edge(strings.Repeat("─", max(0, w-3-ansi.StringWidth(label)))+"╮")
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
	switch m.state {
	case loading:
		return []string{"", " Loading tasks…"}
	case loadFailed:
		return []string{"", " Could not load tasks.", " Press r to retry."}
	}
	if len(m.forest) == 0 {
		return []string{"", " No tasks", " Your workspace is ready."}
	}
	if len(m.rows) == 0 {
		return []string{"", " No matching tasks", " Esc clears filters."}
	}
	type entry struct {
		row, part int
		text      string
	}
	noun := "tasks"
	if len(m.rows) == 1 {
		noun = "task"
	}
	entries := []entry{{row: -1, text: fmt.Sprintf(" %d %s", len(m.rows), noun)}, {row: -1}}
	group := ""
	selectedLine := 0
	for i, row := range m.rows {
		if row.group != group {
			group = row.group
			entries = append(entries, entry{row: -1, text: " " + group})
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
			lines = append(lines, m.paint(entry.text, mutedColor, true))
			continue
		}
		row := m.rows[entry.row]
		task := row.node.Task
		s := ""
		if entry.part == 1 {
			indent := 2 + min(row.depth*2, max(0, w-7))
			if row.depth > 0 {
				indent += 2
			}
			if len(row.node.Children) > 0 {
				indent += 2
			}
			indent = min(indent, max(0, w-18))
			s = strings.Repeat(" ", indent) + terminaltext.Scalar(string(task.Status)) + " · " + task.Priority.String() + fmt.Sprintf(" · %d%%", task.Progress)
		} else {
			context := ""
			if row.context {
				context = "[context] "
			}
			indent := min(row.depth*2, max(0, w-7-len(context)))
			prefix := "  " + strings.Repeat(" ", indent)
			if row.depth*2 > indent {
				prefix = "  … " + strings.Repeat(" ", max(0, indent-2))
			} else if row.depth > 0 {
				prefix += "└ "
			}
			if len(row.node.Children) > 0 {
				if m.collapsed[task.ID] && !m.filter.HasPredicates() && m.dueStart == nil {
					prefix += "▸ "
				} else {
					prefix += "▾ "
				}
			}
			if entry.row == m.selected {
				prefix = ">" + prefix[1:]
			}
			s = prefix + context + terminaltext.Scalar(task.Title)
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
	return lines
}

func (m *Model) prepareFrame() {
	m.rebuildRows()
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
	left := m.panel("Tasks", m.listLines(l.listWidth-2, l.bodyHeight-2), l.listWidth, l.bodyHeight, m.focus == listFocus)
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
	if m.searching {
		header = m.paint(" TUSK  / ", accentColor, true) + terminaltext.Scalar(m.searchDraft) + "▎"
	} else if m.filter.HasPredicates() || m.dueStart != nil {
		header += "  /  Filtered · Esc clears"
		if m.dueLabel != "" {
			header += " · Due " + m.dueLabel
		}
	}
	rows := []string{fitCells(header, l.width)}
	for y := 0; y < l.bodyHeight; y++ {
		rows = append(rows, left[y]+right[y])
	}
	for y, s := range rows {
		fg := textColor
		if m.helpOpen || m.filters != nil {
			s = ansi.Strip(s)
			fg = borderColor
		}
		rows[y] = m.surface(s, fg, canvasColor)
	}
	m.help.Width = l.width - 2
	footer := " " + m.help.ShortHelpView(browseHints())
	if m.focus == detailsFocus {
		footer = " ↑↓ scroll · PgUp/PgDn · Home/End · Tab tasks · ? help · q quit"
	}
	if m.recoveryNeeded {
		footer = "Read outcome unknown · Reload required · q quit"
	} else if m.stale {
		footer = " Stale · r refresh  ·" + footer
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
	if m.stale && !strings.Contains(footer, "Stale") {
		footer = " Stale ·" + footer
	}
	rows = append(rows, m.surface(fitCells(footer, l.width), mutedColor, footerColor))
	if m.helpOpen {
		content := helpLines()
		area := max(0, l.modal.height-4)
		m.helpScroll = min(max(0, m.helpScroll), max(0, len(content)-area))
		rows = m.overlay(rows, l, "Keyboard shortcuts", content, m.helpScroll, "Esc / ? / q close  ·  Ctrl+C exit")
	}
	if m.filters != nil {
		rows = m.overlay(rows, l, "Filter tasks", m.filterLines(l.modal.width-2), 0, "Ctrl+S apply  ·  Esc cancel")
	}
	m.frame = strings.Join(rows, "\n")
}

// overlay reserves the final inner row for controls; only the content scrolls.
func (m *Model) overlay(rows []string, l layout, title string, content []string, offset int, controls string) []string {
	b := l.modal
	area := max(0, b.height-4)
	lines := make([]string, b.height-2)
	for i := 0; i < area; i++ {
		j := max(0, offset) + i
		if j < len(content) {
			lines[i] = " " + content[j]
		}
	}
	lines[len(lines)-1] = " " + controls
	box := m.panel(title, lines, b.width, b.height, true)
	for i, row := range box {
		y := b.y + i
		left := fitCells(cellSlice(rows[y], 0, b.x), b.x)
		right := fitCells(cellSlice(rows[y], b.x+b.width, l.width), l.width-b.x-b.width)
		rows[y] = left + m.surface(row, textColor, dialogColor) + right
	}
	return rows
}
