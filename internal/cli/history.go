package cli

import (
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

func (i *invocation) historyCommand() *cobra.Command {
	c := dataCommand("history <id>", "Show oldest-first task metadata history", cobra.ExactArgs(1))
	c.Example = "  tusk history TASK_ID --json"
	c.RunE = func(c *cobra.Command, args []string) error {
		id, err := inputID(args[0])
		if err != nil {
			return err
		}
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			events, err := s.GetTaskHistory(c.Context(), id)
			return i.queryResult(events, err, jsonFlag(c))
		})
	}
	return c
}
