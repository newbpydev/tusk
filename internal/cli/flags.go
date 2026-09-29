package cli

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

// pflag's integer value selects radix from prefixes. Progress is decimal.
type decimalInt64 int64

func (v *decimalInt64) String() string { return strconv.FormatInt(int64(*v), 10) }
func (*decimalInt64) Type() string     { return "int64" }
func (v *decimalInt64) Set(raw string) error {
	n, err := strconv.ParseInt(raw, 10, 64)
	if err == nil {
		*v = decimalInt64(n)
	}
	return err
}

func textInput(values ...string) error {
	for _, v := range values {
		if !utf8.ValidString(v) || strings.ContainsRune(v, 0) {
			return ports.ErrInvalidText
		}
	}
	return nil
}
func inputID(v string) (string, error) {
	if err := textInput(v); err != nil {
		return "", err
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", core.ErrInvalidTaskID
	}
	return v, nil
}
func inputTags(values []string) ([]string, error) {
	var flat []string
	for _, v := range values {
		if err := textInput(v); err != nil {
			return nil, err
		}
		flat = append(flat, strings.Split(v, ",")...)
	}
	tags, err := core.NormalizeTags(flat)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(tags))
	for n, t := range tags {
		out[n] = string(t)
	}
	return out, nil
}
func dataCommand(use, short string, args cobra.PositionalArgs) *cobra.Command {
	c := &cobra.Command{Use: use, Short: short, Args: syntaxArgs(args)}
	c.Flags().Bool("json", false, "Print complete JSON")
	return c
}
func jsonFlag(c *cobra.Command) bool { v, _ := c.Flags().GetBool("json"); return v }
func (i *invocation) taskResult(task *core.Task, err error, jsonOutput bool) ([]byte, bool, error) {
	if err != nil {
		return nil, false, err
	}
	data, err := i.formatResult(task, jsonOutput)
	return data, true, err
}
