package cli

import (
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

func (i *invocation) treeCommand() *cobra.Command {
	c := dataCommand("tree [id]", "Show the forest or a complete subtree", cobra.MaximumNArgs(1))
	c.Example = "  tusk tree TASK_ID --json"
	c.RunE = func(c *cobra.Command, args []string) error {
		id := ""
		var err error
		if len(args) > 0 {
			id, err = inputID(args[0])
			if err != nil {
				return err
			}
		}
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			nodes, err := s.GetTaskTree(c.Context(), id)
			return i.queryResult(nodes, err, jsonFlag(c))
		})
	}
	return c
}
