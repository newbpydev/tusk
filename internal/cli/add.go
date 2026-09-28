package cli

import (
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

func (i *invocation) addCommand() *cobra.Command {
	c := dataCommand("add <title>", "Create a task", cobra.ExactArgs(1))
	var notes, priority, due, parent string
	var tags []string
	c.Flags().StringVarP(&notes, "notes", "n", "", "Task notes")
	c.Flags().StringVarP(&priority, "priority", "p", "", "low, medium, high, urgent or 1–4")
	c.Flags().StringVarP(&due, "due", "d", "", "Due date expression")
	c.Flags().StringVar(&parent, "parent", "", "Full parent task ID")
	c.Flags().StringArrayVarP(&tags, "tags", "t", nil, "Tags, repeatable and comma-separated")
	c.Example = "  tusk add 'Review release' --due tomorrow --tags work --json"
	c.RunE = func(c *cobra.Command, args []string) error {
		if err := textInput(args[0], notes, due, parent, priority); err != nil {
			return err
		}
		command := ports.CreateTaskCommand{Title: args[0], Description: notes}
		var err error
		if c.Flags().Changed("priority") {
			command.Priority, err = core.ParsePriority(priority)
			if err != nil {
				return err
			}
		}
		if c.Flags().Changed("due") {
			command.Due = &due
		}
		if c.Flags().Changed("parent") {
			parent, err = inputID(parent)
			if err != nil {
				return err
			}
			command.ParentID = &parent
		}
		command.Tags, err = inputTags(tags)
		if err != nil {
			return err
		}
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			task, err := s.CreateTask(c.Context(), command)
			return i.taskResult(task, err, jsonFlag(c))
		})
	}
	return c
}
