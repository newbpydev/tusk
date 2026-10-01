// Docgen regenerates the CLI manuals and static shell scripts from fresh trees.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/newbpydev/tusk/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() { os.Exit(run(os.Args[1:], os.Stderr)) }

func run(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("docgen", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("output", "docs", "Documentation root; replaces only man and completions")
	check := flags.Bool("check", false, "Fail on generated content drift without changing outputs")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *output == "" {
		fmt.Fprintln(stderr, "docgen: expected --output <directory> [--check]")
		return 2
	}
	if err := execute(*output, *check, os.Rename); err != nil {
		fmt.Fprintln(stderr, "docgen:", err)
		return 1
	}
	return 0
}

func execute(output string, check bool, rename func(string, string) error) error {
	if check {
		stage, err := os.MkdirTemp("", "tusk-docgen-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(stage)
		if err := generate(cli.CommandTree(), stage); err != nil {
			return err
		}
		return compare(stage, output)
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}
	for _, path := range []string{output, filepath.Join(output, "man"), filepath.Join(output, "completions")} {
		info, err := os.Lstat(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("output must be a real directory: %s", path)
		}
	}
	stage, err := os.MkdirTemp(output, ".docgen-")
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(stage)
		}
	}()
	if err := generate(cli.CommandTree(), stage); err != nil {
		return err
	}
	if err := publish(stage, output, rename); err != nil {
		cleanup = false
		return fmt.Errorf("promotion failed; recovery files retained at %s: %w", stage, err)
	}
	return nil
}

func generate(root *cobra.Command, stage string) error {
	for _, dir := range []string{"man", "completions"} {
		if err := os.MkdirAll(filepath.Join(stage, dir), 0755); err != nil {
			return err
		}
	}
	// This fixed documentation epoch is a source constant, never the build clock.
	date := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	header := &doc.GenManHeader{Section: "1", Date: &date, Source: "Tusk", Manual: "Tusk Manual"}
	escapeSynopsis(root)
	if err := doc.GenManTreeFromOpts(root, doc.GenManTreeOptions{Header: header, Path: filepath.Join(stage, "man"), CommandSeparator: "-"}); err != nil {
		return err
	}
	for _, shell := range []string{"bash", "zsh", "fish"} {
		var out, errs bytes.Buffer
		if code := cli.Run(context.Background(), []string{"completion", shell}, cli.Options{Stdout: &out, Stderr: &errs}); code != 0 {
			return fmt.Errorf("completion generation failed: %s", errs.String())
		}
		if err := os.WriteFile(filepath.Join(stage, "completions", "tusk."+shell), out.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

// md2man parses angle-bracket arguments as HTML unless Markdown-escaped.
// The detached documentation tree keeps the runtime command grammar intact.
func escapeSynopsis(root *cobra.Command) {
	root.Use = strings.NewReplacer("<", `\<`, ">", `\>`).Replace(root.Use)
	for _, child := range root.Commands() {
		escapeSynopsis(child)
	}
}

func compare(stage, output string) error {
	for _, group := range []string{"man", "completions"} {
		left, err := readGroup(filepath.Join(stage, group))
		if err != nil {
			return err
		}
		right, err := readGroup(filepath.Join(output, group))
		if err != nil {
			return err
		}
		if len(left) != len(right) {
			return fmt.Errorf("generated member drift: %s", group)
		}
		for name, data := range left {
			other, ok := right[name]
			if !ok || !bytes.Equal(data, other) {
				return fmt.Errorf("generated content drift: %s/%s", group, name)
			}
		}
	}
	return nil
}

func readGroup(root string) (map[string][]byte, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			return nil, fmt.Errorf("nonregular generated member: %s", entry.Name())
		}
		data, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		files[entry.Name()] = data
	}
	return files, nil
}

func publish(stage, output string, rename func(string, string) error) error {
	var backed, installed []string
	rollback := func(cause error) error {
		for _, group := range installed {
			cause = errors.Join(cause, os.RemoveAll(filepath.Join(output, group)))
		}
		for _, group := range backed {
			cause = errors.Join(cause, rename(filepath.Join(stage, "previous-"+group), filepath.Join(output, group)))
		}
		return cause
	}
	for _, group := range []string{"man", "completions"} {
		path := filepath.Join(output, group)
		if _, err := os.Lstat(path); err == nil {
			if err := rename(path, filepath.Join(stage, "previous-"+group)); err != nil {
				return rollback(err)
			}
			backed = append(backed, group)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return rollback(err)
		}
		if err := rename(filepath.Join(stage, group), path); err != nil {
			return rollback(err)
		}
		installed = append(installed, group)
	}
	return nil
}
