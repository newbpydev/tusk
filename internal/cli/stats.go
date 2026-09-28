package cli

import (
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

func (i *invocation) statsCommand() *cobra.Command {
	c := dataCommand("stats", "Show retained-task statistics", cobra.NoArgs)
	c.Example = "  tusk stats --json"
	c.RunE = func(c *cobra.Command, _ []string) error {
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			stats, err := s.GetStats(c.Context())
			return i.queryResult(stats, err, jsonFlag(c))
		})
	}
	return c
}
