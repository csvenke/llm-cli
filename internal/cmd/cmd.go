package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	askcmd "llm/internal/cmd/ask"
	commitcmd "llm/internal/cmd/commit"
	ghcmd "llm/internal/cmd/gh"
	prcmd "llm/internal/cmd/gh/pr"
	"llm/internal/gh"
	"llm/internal/git"
	"llm/internal/providers"
)

var ErrUnknownCommand = errors.New("unknown command")

type Dependencies struct {
	Provider providers.Provider
	Stdout   io.Writer
	Stderr   io.Writer
	Git      git.Client
	GH       gh.Client
}

type Handler func(ctx context.Context, deps Dependencies, args []string) error

type Command struct {
	Name        string
	Usage       string
	Description string
	Run         Handler
	Subcommands []*Command
}

type Registry struct {
	commands []*Command
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		commands: []*Command{
			{
				Name:        askcmd.Name,
				Usage:       askcmd.Usage,
				Description: askcmd.Description,
				Run: func(ctx context.Context, deps Dependencies, args []string) error {
					return askcmd.Run(ctx, deps.Provider, deps.Stdout, deps.Stderr, args)
				},
			},
			{
				Name:        commitcmd.Name,
				Usage:       commitcmd.Usage,
				Description: commitcmd.Description,
				Run: func(ctx context.Context, deps Dependencies, args []string) error {
					return commitcmd.Run(ctx, deps.Provider, deps.Git, deps.Stderr, args)
				},
			},
			{
				Name:        ghcmd.Name,
				Usage:       ghcmd.Usage,
				Description: ghcmd.Description,
				Subcommands: []*Command{
					{
						Name:        prcmd.Name,
						Usage:       prcmd.Usage,
						Description: prcmd.Description,
						Run: func(ctx context.Context, deps Dependencies, args []string) error {
							return prcmd.Run(ctx, deps.Provider, deps.GH, deps.Stdout, deps.Stderr, args)
						},
					},
				},
			},
		},
	}
}

func UsageTo(w io.Writer) {
	defaultRegistry.UsageTo(w)
}

func Run(ctx context.Context, provider providers.Provider, stdout, stderr io.Writer, args []string) error {
	deps := Dependencies{
		Provider: provider,
		Stdout:   stdout,
		Stderr:   stderr,
		Git:      &git.RealClient{},
		GH:       &gh.RealClient{},
	}

	return defaultRegistry.Run(ctx, deps, args)
}

func (r *Registry) UsageTo(w io.Writer) {
	fmt.Fprintf(w, "Usage: llm <command> [options]\n\n")
	fmt.Fprintf(w, "Commands:\n")
	r.writeCommands(w, r.commands, 2)
}

func (r *Registry) Run(ctx context.Context, deps Dependencies, args []string) error {
	if len(args) == 0 {
		return ErrUnknownCommand
	}

	deps = normalizeDependencies(deps)

	cmd := findCommand(r.commands, args[0])
	if cmd == nil {
		return ErrUnknownCommand
	}

	return runCommand(ctx, deps, cmd, cmd.Name, args[1:])
}

func runCommand(ctx context.Context, deps Dependencies, cmd *Command, path string, args []string) error {
	if len(cmd.Subcommands) > 0 {
		if len(args) == 0 {
			return fmt.Errorf("usage: llm %s <subcommand>", path)
		}

		sub := findCommand(cmd.Subcommands, args[0])
		if sub == nil {
			return fmt.Errorf("unknown %s subcommand %q (usage: llm %s <subcommand>)", cmd.Name, args[0], path)
		}

		return runCommand(ctx, deps, sub, path+" "+sub.Name, args[1:])
	}

	if cmd.Run == nil {
		return nil
	}

	return cmd.Run(ctx, deps, args)
}

func findCommand(commands []*Command, name string) *Command {
	for _, cmd := range commands {
		if cmd.Name == name {
			return cmd
		}
	}

	return nil
}

func (r *Registry) writeCommands(w io.Writer, commands []*Command, indent int) {
	for _, cmd := range commands {
		if indent <= 2 {
			fmt.Fprintf(w, "%s%-19s %s\n", spaces(indent), cmd.Usage, cmd.Description)
		} else {
			fmt.Fprintf(w, "%s%-17s %s\n", spaces(indent), cmd.Usage, cmd.Description)
		}

		if len(cmd.Subcommands) > 0 {
			r.writeCommands(w, cmd.Subcommands, indent+2)
		}
	}
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}

	return fmt.Sprintf("%*s", n, "")
}

func normalizeDependencies(deps Dependencies) Dependencies {
	if deps.Stdout == nil {
		deps.Stdout = io.Discard
	}

	if deps.Stderr == nil {
		deps.Stderr = io.Discard
	}

	if deps.Git == nil {
		deps.Git = &git.RealClient{}
	}

	if deps.GH == nil {
		deps.GH = &gh.RealClient{}
	}

	return deps
}
