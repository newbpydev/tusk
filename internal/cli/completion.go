package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandTree returns a fresh tree for static documentation generation.
// Execute application invocations through Run, which owns streams and outcomes.
// Construction does not read configuration or discover storage or terminals.
func CommandTree() *cobra.Command {
	return newInvocation(Options{}).root
}

func (i *invocation) completionCommand() *cobra.Command {
	c := &cobra.Command{
		Use: "completion <bash|zsh|fish>", Short: "Generate a shell completion script",
		Args: syntaxArgs(cobra.ExactArgs(1)), ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(_ *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return i.root.GenBashCompletionV2(&i.output, true)
			case "zsh":
				return i.root.GenZshCompletion(&i.output)
			case "fish":
				return i.root.GenFishCompletion(&i.output, true)
			default:
				return syntaxError{}
			}
		},
	}
	return c
}

func staticCompletions(c *cobra.Command, registerFlags bool) {
	c.DisableAutoGenTag = true
	if len(c.ValidArgs) == 0 {
		c.ValidArgsFunction = cobra.NoFileCompletions
	}
	for _, child := range c.Commands() {
		if registerFlags && child.Flags().Lookup("status") != nil {
			_ = child.RegisterFlagCompletionFunc("status", cobra.FixedCompletions([]string{"todo", "in-progress", "blocked", "done"}, cobra.ShellCompDirectiveNoFileComp))
		}
		if registerFlags && child.Flags().Lookup("priority") != nil {
			_ = child.RegisterFlagCompletionFunc("priority", cobra.FixedCompletions([]string{"low", "medium", "high", "urgent", "1", "2", "3", "4"}, cobra.ShellCompDirectiveNoFileComp))
		}
		// Non-enum flag values must not ask storage or propose task data.
		if registerFlags {
			child.Flags().VisitAll(func(flag *pflag.Flag) {
				if flag.Name != "status" && flag.Name != "priority" {
					_ = child.RegisterFlagCompletionFunc(flag.Name, cobra.NoFileCompletions)
				}
			})
		}
		staticCompletions(child, registerFlags)
	}
}
