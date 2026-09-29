package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/rivo/uniseg"
)

// wrapCells never splits a grapheme. Oversize clusters become visible ASCII
// code-point escapes, which can themselves safely wrap at a one-cell width.
func wrapCells(text string, width int) string {
	var out strings.Builder
	used := 0
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		cluster := g.Str()
		cells := ansi.StringWidth(cluster)
		if cells > width {
			var escaped strings.Builder
			for _, r := range cluster {
				fmt.Fprintf(&escaped, `\U%08x`, r)
			}
			for _, r := range escaped.String() {
				if used == width {
					out.WriteByte('\n')
					used = 0
				}
				out.WriteRune(r)
				used++
			}
			continue
		}
		if used+cells > width {
			out.WriteByte('\n')
			used = 0
		}
		out.WriteString(cluster)
		used += cells
	}
	return out.String()
}
func pad(text string, width int) string {
	return text + strings.Repeat(" ", max(0, width-ansi.StringWidth(text)))
}
func (f *formatter) status(s core.Status) string {
	text := sanitize(string(s))
	if f.renderer == nil {
		return text
	}
	color := "4"
	switch s {
	case core.StatusDone:
		color = "2"
	case core.StatusBlocked:
		color = "1"
	case core.StatusInProgress:
		color = "3"
	}
	return f.renderer.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(text)
}
func (f *formatter) line(text string) string {
	if f.terminal {
		return wrapCells(text, f.width) + "\n"
	}
	return text + "\n"
}
func (f *formatter) tasks(tasks []core.Task) string {
	const header = "ID\tSTATUS\tPRIORITY\tPROGRESS\tDUE\tTITLE\n"
	var out strings.Builder
	if !f.terminal {
		out.WriteString(header)
		for _, t := range tasks {
			due := "-"
			if t.DueDate != nil {
				due = t.DueDate.UTC().Format(time.RFC3339Nano)
			}
			fmt.Fprintf(&out, "%s\t%s\t%s\t%d%%\t%s\t%s\n", sanitize(t.ID), sanitize(string(t.Status)), t.Priority.String(), t.Progress, due, sanitize(t.Title))
		}
		return out.String()
	}
	idWidth := 36
	for _, t := range tasks {
		idWidth = max(idWidth, ansi.StringWidth(sanitize(t.ID)))
	}
	titleWidth := f.width - idWidth - 11 - 6 - 8 - 10 - 10
	if titleWidth >= 12 {
		row := func(id, status, priority, progress, due, title string) string {
			return pad(id, idWidth) + "  " + pad(status, 11) + "  " + pad(priority, 6) + "  " + pad(progress, 8) + "  " + pad(due, 10) + "  " + ansi.Truncate(title, titleWidth, "…") + "\n"
		}
		out.WriteString(row("ID", "STATUS", "PRIO", "PROGRESS", "DUE", "TITLE"))
		for _, t := range tasks {
			due := "-"
			if t.DueDate != nil {
				due = t.DueDate.In(f.location).Format("2006-01-02")
			}
			out.WriteString(row(sanitize(t.ID), f.status(t.Status), t.Priority.String(), fmt.Sprintf("%d%%", t.Progress), due, sanitize(t.Title)))
		}
	} else {
		if len(tasks) == 0 {
			return f.line("ID STATUS PRIORITY PROGRESS DUE TITLE")
		}
		for n, t := range tasks {
			if n > 0 {
				out.WriteByte('\n')
			}
			due := "-"
			if t.DueDate != nil {
				due = t.DueDate.In(f.location).Format("2006-01-02")
			}
			out.WriteString(f.line(sanitize(t.ID)))
			// Style complete wrapped status lines, never split an ANSI sequence.
			statusLine := wrapCells("Status: "+sanitize(string(t.Status)), f.width)
			if f.renderer != nil {
				statusLine = f.renderer.NewStyle().Bold(true).Render(statusLine)
			}
			out.WriteString(statusLine + "\n")
			for _, field := range []string{"Priority: " + t.Priority.String(), fmt.Sprintf("Progress: %d%%", t.Progress), "Due: " + due, "Title: " + sanitize(t.Title)} {
				out.WriteString(f.line(field))
			}
		}
	}
	return out.String()
}
func (f *formatter) tree(nodes []*core.TaskNode, prefix string) (string, error) {
	var out strings.Builder
	for n, node := range nodes {
		if node == nil {
			return "", ports.ErrInvalidRecord
		}
		last := n == len(nodes)-1
		branch := "|-- "
		next := "|   "
		if last {
			branch = "`-- "
			next = "    "
		}
		text := fmt.Sprintf("%s [%s %d%%] %s (depth %d)", sanitize(node.Task.ID), sanitize(string(node.Task.Status)), node.Task.Progress, sanitize(node.Task.Title), node.Depth)
		if f.terminal {
			indent := min(len(prefix)+len(branch), f.width-1)
			for n, line := range strings.Split(wrapCells(text, f.width-indent), "\n") {
				lead := strings.Repeat(" ", indent)
				if n == 0 {
					lead = ansi.Truncate(prefix+branch, indent, "")
				}
				out.WriteString(lead + line + "\n")
			}
		} else {
			out.WriteString(prefix + branch + text + "\n")
		}
		children, err := f.tree(node.Children, prefix+next)
		if err != nil {
			return "", err
		}
		out.WriteString(children)
	}
	return out.String(), nil
}
func (f *formatter) format(value any) ([]byte, error) {
	var out string
	switch v := value.(type) {
	case *core.Task:
		if v == nil {
			return nil, ports.ErrInvalidRecord
		}
		out = f.tasks([]core.Task{*v})
	case []core.Task:
		out = f.tasks(v)
	case []*core.TaskNode:
		if len(v) == 0 {
			out = f.line("No tasks.")
		} else {
			var err error
			out, err = f.tree(v, "")
			if err != nil {
				return nil, err
			}
		}
	case ports.DeleteResult:
		out = f.line(fmt.Sprintf("Deleted %d task(s): %s", v.DeletedCount, sanitize(strings.Join(sortedStrings(v.DeletedIDs), ", "))))
	case ports.TaskStats:
		labels := []string{"Total", "todo", "in-progress", "blocked", "done", "Done", "Completion percent", "Overdue", "Completed last 7 days"}
		values := []int{v.Total, v.ByStatus[core.StatusTodo], v.ByStatus[core.StatusInProgress], v.ByStatus[core.StatusBlocked], v.ByStatus[core.StatusDone], v.Done, v.CompletionPercent, v.Overdue, v.CompletedLast7Days}
		for n, label := range labels {
			sep := "\t"
			if f.terminal {
				sep = ": "
			}
			out += f.line(label + sep + strconv.Itoa(values[n]))
		}
	case []ports.TaskEvent:
		out = f.history(v)
	default:
		return nil, ports.ErrInvalidRecord
	}
	return []byte(out), nil
}

func (f *formatter) history(events []ports.TaskEvent) string {
	headers := []string{"SEQUENCE", "OCCURRED_AT", "KIND", "CHANGED_FIELDS"}
	rows := make([][]string, len(events))
	widths := []int{8, 11, 4, 14}
	for n, e := range events {
		rows[n] = []string{strconv.FormatInt(e.Sequence, 10), e.OccurredAt.UTC().Format(time.RFC3339Nano), sanitize(string(e.Kind)), sanitize(strings.Join(sortedStrings(e.ChangedFields), ","))}
		for col, value := range rows[n] {
			widths[col] = max(widths[col], ansi.StringWidth(value))
		}
	}
	var out strings.Builder
	if !f.terminal {
		out.WriteString(strings.Join(headers, "\t") + "\n")
		for _, row := range rows {
			out.WriteString(strings.Join(row, "\t") + "\n")
		}
	} else if widths[0]+widths[1]+widths[2]+widths[3]+6 <= f.width {
		writeRow := func(row []string) {
			for col, value := range row {
				if col > 0 {
					out.WriteString("  ")
				}
				if col == 3 {
					out.WriteString(value)
				} else {
					out.WriteString(pad(value, widths[col]))
				}
			}
			out.WriteByte('\n')
		}
		writeRow(headers)
		for _, row := range rows {
			writeRow(row)
		}
	} else {
		if len(rows) == 0 {
			return f.line(strings.Join(headers, " "))
		}
		labels := []string{"Sequence", "Occurred at", "Kind", "Changed fields"}
		for n, row := range rows {
			if n > 0 {
				out.WriteByte('\n')
			}
			for col, value := range row {
				out.WriteString(f.line(labels[col] + ": " + value))
			}
		}
	}
	return out.String()
}
