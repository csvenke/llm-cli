package ask

import (
	"context"
	"fmt"
	"io"
	"strings"

	"llm/internal/loading"
	"llm/internal/providers"
)

// Run executes the ask command with the given arguments.
// Returns an error if no arguments are provided or if the provider fails.
func Run(ctx context.Context, provider providers.Provider, output io.Writer, stderr io.Writer, args []string) error {
	if output == nil {
		output = io.Discard
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: llm ask <question>")
	}

	question := strings.Join(args, " ")

	ind := loading.Start(stderr)
	response, err := provider.Complete(ctx, "", question)
	ind.Stop()

	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(output, response)
	return err
}
