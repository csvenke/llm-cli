package ask

import (
	"context"
	"fmt"
	"strings"

	"llm/internal/providers"
)

func Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: llm ask <question>")
	}

	provider, err := providers.ResolveByAPIKey()
	if err != nil {
		return err
	}

	question := strings.Join(args, " ")
	response, err := provider.Complete(ctx, "", question)
	if err != nil {
		return err
	}

	fmt.Println(response)
	return nil
}
