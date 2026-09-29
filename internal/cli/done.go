package cli

import (
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

func (i *invocation) doneCommand() *cobra.Command {
	c := dataCommand("done <id>", "Complete a task and its descendants", cobra.ExactArgs(1))
	c.Example = "  tusk done TASK_ID --json"
	c.RunE = func(c *cobra.Command, args []string) error {
		id, err := inputID(args[0])
		if err != nil {
			return err
		}
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			task, err := s.CompleteTask(c.Context(), ports.TaskCommand{ID: id})
			return i.taskResult(task, err, jsonFlag(c))
		})
	}
	return c
}
