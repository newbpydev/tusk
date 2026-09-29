package tui

import (
	"slices"
	"strings"
	"time"

	"github.com/newbpydev/tusk/internal/core"
)

type taskRow struct {
	node    *core.TaskNode
	depth   int
	group   string
	context bool
}

// project derives visible rows from the complete accepted forest. Filtering
// retains ancestor paths, while collapse is a session preference left intact
// during search. Detached sibling slices keep service snapshots immutable.
func project(forest []*core.TaskNode, filter core.TaskFilter, start, end *time.Time, collapsed map[string]bool, now time.Time, location *time.Location) []taskRow {
	filter.SearchTerm = strings.TrimSpace(filter.SearchTerm)
	filtered := filter.HasPredicates() || start != nil || end != nil
	matches, included := map[string]bool{}, map[string]bool{}
	if filtered {
		var tasks []core.Task
		var collect func([]*core.TaskNode)
		collect = func(nodes []*core.TaskNode) {
			for _, n := range nodes {
				tasks = append(tasks, n.Task)
				collect(n.Children)
			}
		}
		collect(forest)
		for _, t := range core.FilterTasks(tasks, filter) {
			if start != nil && (t.DueDate == nil || t.DueDate.Before(*start)) {
				continue
			}
			if end != nil && (t.DueDate == nil || !t.DueDate.Before(*end)) {
				continue
			}
			matches[t.ID] = true
		}
		var include func(*core.TaskNode) bool
		include = func(n *core.TaskNode) bool {
			yes := matches[n.Task.ID]
			for _, c := range n.Children {
				if include(c) {
					yes = true
				}
			}
			included[n.Task.ID] = yes
			return yes
		}
		for _, n := range forest {
			include(n)
		}
	}
	local := now.In(location)
	nextDay := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, location)
	order := []core.SortOrder{{Field: core.SortByPriority, Direction: core.SortDesc}, {Field: core.SortByDueDate, Direction: core.SortAsc}, {Field: core.SortByCreatedAt, Direction: core.SortAsc}}
	sorted := func(nodes []*core.TaskNode) []*core.TaskNode {
		out := append([]*core.TaskNode(nil), nodes...)
		slices.SortStableFunc(out, func(a, b *core.TaskNode) int { return core.CompareTasks(&a.Task, &b.Task, order) })
		return out
	}
	groups := make([][]*core.TaskNode, 4)
	for _, n := range sorted(forest) {
		group := 2
		switch {
		case n.Task.Status == core.StatusDone:
			group = 3
		case n.Task.DueDate != nil:
			group = 1
			if n.Task.DueDate.Before(nextDay) {
				group = 0
			}
		}
		groups[group] = append(groups[group], n)
	}
	var rows []taskRow
	var visit func(*core.TaskNode, int, string)
	visit = func(n *core.TaskNode, depth int, group string) {
		if filtered && !included[n.Task.ID] {
			return
		}
		rows = append(rows, taskRow{n, depth, group, filtered && !matches[n.Task.ID]})
		if !filtered && collapsed[n.Task.ID] {
			return
		}
		for _, child := range sorted(n.Children) {
			visit(child, depth+1, group)
		}
	}
	for i, name := range []string{"Today", "Upcoming", "Backlog", "Completed"} {
		for _, n := range groups[i] {
			visit(n, 0, name)
		}
	}
	return rows
}
