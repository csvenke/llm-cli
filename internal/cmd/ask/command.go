package askcmd

import (
	"context"
	"io"

	"llm/internal/ask"
	"llm/internal/providers"
)

const (
	Name        = "ask"
	Usage       = "ask <question>"
	Description = "Ask a question"
)

var RunFunc = ask.Run

func Run(ctx context.Context, provider providers.Provider, stdout, stderr io.Writer, args []string) error {
	return RunFunc(ctx, provider, stdout, stderr, args)
}
