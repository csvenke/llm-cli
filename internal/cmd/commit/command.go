package commitcmd

import (
	"context"
	"io"

	"llm/internal/commit"
	"llm/internal/git"
	"llm/internal/providers"
)

const (
	Name        = "commit"
	Usage       = "commit [-a|--amend]"
	Description = "Draft a commit message"
)

var RunFunc = commit.Run

func Run(ctx context.Context, provider providers.Provider, gitClient git.Client, stderr io.Writer, args []string) error {
	return RunFunc(ctx, provider, gitClient, stderr, args)
}
