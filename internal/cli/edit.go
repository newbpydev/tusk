package cli

import (
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

func (i *invocation) editCommand() *cobra.Command {
	c := dataCommand("edit <id>", "Patch supplied task fields", cobra.ExactArgs(1))
	var title, notes, priority, status, due, parent string
	var tags []string
	var progress decimalInt64
	var clearDue, clearTags, root bool
	f := c.Flags()
	f.StringVar(&title, "title", "", "Task title")
	f.StringVarP(&notes, "notes", "n", "", "Task notes (empty clears)")
	f.StringVarP(&priority, "priority", "p", "", "Priority")
	f.StringVarP(&status, "status", "s", "", "Task status")
	f.Var(&progress, "progress", "Manual leaf progress 0–100")
	f.StringVarP(&due, "due", "d", "", "Due date expression")
	f.BoolVar(&clearDue, "clear-due", false, "Clear due date")
	f.StringArrayVarP(&tags, "tags", "t", nil, "Repeatable comma-separated tags")
	f.BoolVar(&clearTags, "clear-tags", false, "Clear all tags")
	f.StringVar(&parent, "parent", "", "Full parent task ID")
	f.BoolVar(&root, "root", false, "Move task to the root")
	c.Example = "  tusk edit TASK_ID --notes '' --clear-due --json"
	c.Args = func(c *cobra.Command, args []string) error {
		if err := syntaxArgs(cobra.ExactArgs(1))(c, args); err != nil {
			return err
		}
		changed := false
		for _, name := range []string{"title", "notes", "priority", "status", "progress", "due", "tags", "parent"} {
			changed = changed || f.Changed(name)
		}
		if !(changed || clearDue || clearTags || root) || f.Changed("due") && clearDue || f.Changed("tags") && clearTags || f.Changed("parent") && root || f.Changed("progress") && (f.Changed("status") || f.Changed("parent") || root) {
			return syntaxError{ports.ErrInvalidCommand}
		}
		return nil
	}
	c.RunE = func(c *cobra.Command, args []string) error {
		id, err := inputID(args[0])
		if err != nil {
			return err
		}
		if err = textInput(title, notes, priority, status, due, parent); err != nil {
			return err
		}
		command := ports.UpdateTaskCommand{ID: id, ClearDue: clearDue, ClearParent: root}
		if f.Changed("title") {
			command.Title = &title
		}
		if f.Changed("notes") {
			command.Description = &notes
		}
		if f.Changed("priority") {
			v, e := core.ParsePriority(priority)
			if e != nil {
				return e
			}
			command.Priority = &v
		}
		if f.Changed("status") {
			v, e := core.ParseStatus(status)
			if e != nil {
				return e
			}
			command.Status = &v
		}
		if f.Changed("progress") {
			if progress < 0 || progress > 100 {
				return core.ErrInvalidProgress
			}
			v := int(progress)
			command.Progress = &v
		}
		if f.Changed("due") {
			command.Due = &due
		}
		if f.Changed("parent") {
			parent, err = inputID(parent)
			if err != nil {
				return err
			}
			command.ParentID = &parent
		}
		if f.Changed("tags") || clearTags {
			v, e := inputTags(tags)
			if e != nil {
				return e
			}
			command.Tags = &v
		}
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			task, err := s.UpdateTask(c.Context(), command)
			return i.taskResult(task, err, jsonFlag(c))
		})
	}
	return c
}
