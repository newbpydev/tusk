package cli

import (
	"context"
	"fmt"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

const errForceRequired core.Error = "noninteractive deletion requires --force; deleting descendants also requires --recursive"

func (i *invocation) deleteCommand() *cobra.Command {
	c := dataCommand("delete <id>", "Delete a task with explicit consent", cobra.ExactArgs(1))
	var force, recursive bool
	c.Flags().BoolVar(&force, "force", false, "Skip confirmation (does not imply recursion)")
	c.Flags().BoolVar(&recursive, "recursive", false, "Delete the whole subtree")
	c.Example = "  tusk delete TASK_ID --force --recursive --json"
	c.RunE = func(c *cobra.Command, args []string) error {
		id, err := inputID(args[0])
		if err != nil {
			return err
		}
		if !force {
			if jsonFlag(c) {
				return errForceRequired
			}
			facts := i.terminal()
			if !facts.In || !facts.Out || !facts.Err {
				return errForceRequired
			}
			if i.options.Confirm == nil {
				return ports.ErrInvalidServiceOptions
			}
		}
		return i.invoke(c, func(s ports.TaskService) ([]byte, bool, error) {
			command := ports.DeleteTaskCommand{ID: id, Force: force, Recursive: recursive}
			if !force {
				preview, err := s.PreviewDeleteTask(c.Context(), id)
				if err != nil {
					return nil, false, err
				}
				if len(preview.IDs) > 1 && !recursive {
					return nil, false, ports.ErrChildrenPresent
				}
				prompt := fmt.Sprintf("%s (%s)\nDelete %d task(s)? [y/N] ", sanitize(preview.Target.Title), sanitize(preview.Target.ID), len(preview.IDs))
				if err = writeAll(i.options.Stderr, []byte(prompt)); err != nil {
					return nil, false, err
				}
				yes, err := i.options.Confirm(c.Context())
				if err != nil {
					return nil, false, err
				}
				if err = c.Context().Err(); err != nil {
					return nil, false, err
				}
				if !yes {
					return nil, false, writeAll(i.options.Stderr, []byte("Deletion canceled.\n"))
				}
				command.Expected = &preview
			}
			result, err := s.DeleteTask(c.Context(), command)
			if err != nil {
				return nil, false, err
			}
			data, err := i.formatResult(result, jsonFlag(c))
			return data, true, err
		})
	}
	return c
}

// Confirmation never acquires storage; the composition root owns input lifetime.
type Confirmation func(context.Context) (bool, error)
