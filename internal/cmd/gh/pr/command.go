package prcmd

import (
	"context"
	"fmt"
	"io"

	"llm/internal/gh"
	"llm/internal/providers"
)

const (
	Name        = "pr"
	Usage       = "pr"
	Description = "Create a GitHub pull request"
)

var (
	RunFunc               = gh.Run
	CreatePullRequestFunc = gh.CreatePullRequest
)

func Run(ctx context.Context, provider providers.Provider, client gh.Client, stdout, stderr io.Writer, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: llm gh pr")
	}

	pr, err := RunFunc(ctx, provider, client, stderr, args)
	if err != nil {
		return err
	}

	output, err := CreatePullRequestFunc(pr)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, output)
	return err
}
