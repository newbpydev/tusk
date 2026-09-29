package cli

import (
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

func (i *invocation) listCommand() *cobra.Command {
	c := dataCommand("list", "List tasks with composable filters", cobra.NoArgs)
	var statuses, priorities, tags []string
	var due, search, parent string
	var all, root bool
	f := c.Flags()
	f.StringArrayVarP(&statuses, "status", "s", nil, "Repeatable status (OR)")
	f.StringArrayVarP(&priorities, "priority", "p", nil, "Repeatable priority (OR)")
	f.StringArrayVarP(&tags, "tags", "t", nil, "Require all tags")
	f.StringVar(&due, "due", "", "Local civil day")
	f.StringVar(&search, "search", "", "Literal title/notes search")
	f.StringVar(&parent, "parent", "", "Full parent task ID")
	f.BoolVar(&root, "root", false, "Root tasks only")
	f.BoolVar(&all, "all", false, "Include all statuses")
	c.Example = "  tusk list --status todo --status blocked --tags work --json"
	c.Args = func(c *cobra.Command, args []string) error {
		if err := syntaxArgs(cobra.NoArgs)(c, args); err != nil {
			return err
		}
		if all && f.Changed("status") || root && f.Changed("parent") {
			return syntaxError{ports.ErrInvalidCommand}
		}
		return nil
	}
	c.RunE = func(c *cobra.Command, _ []string) error {
		if err := textInput(due, search, parent); err != nil {
			return err
		}
		q := ports.TaskQuery{All: all, Filter: core.TaskFilter{RootOnly: root, SearchTerm: search}}
		for _, value := range statuses {
			status, err := core.ParseStatus(value)
			if err != nil {
				return err
			}
			q.Filter.Statuses = append(q.Filter.Statuses, status)
		}
		for _, value := range priorities {
			priority, err := core.ParsePriority(value)
			if err != nil {
				return err
			}
			q.Filter.Priorities = append(q.Filter.Priorities, priority)
		}
		normalized, err := inputTags(tags)
		if err != nil {
			return err
		}
		for _, tag := range normalized {
			q.Filter.Tags = append(q.Filter.Tags, core.Tag(tag))
		}
		if f.Changed("due") {
			q.Due = &due
		}
		if f.Changed("parent") {
			parent, err = inputID(parent)
			if err != nil {
				return err
			}
			q.Filter.ParentID = &parent
		}
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			tasks, err := s.ListTasks(c.Context(), q)
			return i.queryResult(tasks, err, jsonFlag(c))
		})
	}
	return c
}
