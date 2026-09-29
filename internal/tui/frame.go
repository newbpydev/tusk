package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
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
	lines := []string{m.paint(fmt.Sprintf(" %d task roots", len(m.forest)), mutedColor, false), ""}
	var visit func([]*core.TaskNode, int)
	visit = func(nodes []*core.TaskNode, depth int) {
		for _, n := range nodes {
			if len(lines) >= h {
				return
			}
			indent := min(depth*2, max(0, w-7))
			prefix := "  " + strings.Repeat(" ", indent)
			if depth > 0 {
				prefix += "└ "
			}
			if depth*2 > indent {
				prefix = "  … " + strings.Repeat(" ", max(0, indent-2))
			}
			selected := n == m.forest[0]
			if selected {
				prefix = ">" + prefix[1:]
			}
			title := titleCells(prefix+terminaltext.Scalar(n.Task.Title), w)
			meta := titleCells("  "+terminaltext.Scalar(string(n.Task.Status))+" · "+n.Task.Priority.String()+fmt.Sprintf(" · %d%%", n.Task.Progress), w)
			if selected {
				title = m.surface(m.paint(title, textColor, true), textColor, selectionColor)
				meta = m.surface(meta, accentColor, selectionColor)
			} else {
				meta = m.paint(meta, mutedColor, false)
			}
			lines = append(lines, title, meta)
			visit(n.Children, depth+1)
		}
	}
	visit(m.forest, 0)
	return lines
}

func (m *Model) detailLines(w int) []string {
	if len(m.forest) == 0 {
		return []string{"", " Select a task to see its details."}
	}
	t := m.forest[0].Task
	lines := []string{"", "  " + m.paint(titleCells(terminaltext.Scalar(t.Title), max(0, w-4)), textColor, true), "",
		"  Status      " + terminaltext.Scalar(string(t.Status)), "  Priority    " + t.Priority.String(), fmt.Sprintf("  Progress    %d%%", t.Progress), "", "  " + m.paint("Notes", textColor, true), ""}
	for _, line := range strings.Split(terminaltext.Multiline(t.Description), "\n") {
		lines = append(lines, "  "+titleCells(line, max(0, w-4)))
	}
	if m.notes.Focused() {
		lines = append(lines, strings.Split(m.notes.View(), "\n")...)
	}
	return lines
}

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
	left := m.panel("Tasks", m.listLines(l.listWidth-2, l.bodyHeight-2), l.listWidth, l.bodyHeight, m.focus == listFocus)
	right := m.panel("Task details", m.detailLines(l.detailsWidth-2), l.detailsWidth, l.bodyHeight, m.focus == detailsFocus)
	header := m.paint(" TUSK", accentColor, true) + m.paint("  /  Personal workspace", mutedColor, false)
	rows := []string{fitCells(header, l.width)}
	for y := 0; y < l.bodyHeight; y++ {
		rows = append(rows, left[y]+right[y])
	}
	for y, s := range rows {
		fg := textColor
		if m.helpOpen {
			s = ansi.Strip(s)
			fg = borderColor
		}
		rows[y] = m.surface(s, fg, canvasColor)
	}
	m.help.Width = l.width - 2
	footer := " " + m.help.ShortHelpView(browseHints())
	if m.state == loadFailed {
		footer = " r retry  ·" + footer
	}
	if m.helpOpen {
		footer = " ↑↓ scroll help  ·  Esc / ? / q close  ·  Ctrl+C exit"
	}
	rows = append(rows, m.surface(fitCells(footer, l.width), mutedColor, footerColor))
	if m.helpOpen {
		content := helpLines()
		area := max(0, l.modal.height-4)
		m.helpScroll = min(max(0, m.helpScroll), max(0, len(content)-area))
		rows = m.overlay(rows, l, "Keyboard shortcuts", content, m.helpScroll, "Esc / ? / q close  ·  Ctrl+C exit")
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
