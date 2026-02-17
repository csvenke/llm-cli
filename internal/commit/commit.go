package commit

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"time"

	"llm/internal/git"
	"llm/internal/providers"
)

//go:embed prompt.md
var systemPrompt string

func Run(ctx context.Context, args []string) error {
	var amend bool
	commitFlags := flag.NewFlagSet("commit", flag.ExitOnError)
	commitFlags.BoolVar(&amend, "a", false, "amend the last commit")
	commitFlags.BoolVar(&amend, "amend", false, "amend the last commit")
	if err := commitFlags.Parse(args); err != nil {
		return err
	}

	diffArgs := []string{"diff", "--staged"}
	if amend {
		parent, err := git.Exec("rev-parse", "--verify", "HEAD~1")
		if err != nil {
			// Initial commit — create an empty tree object to diff against
			emptyTree, treeErr := git.Exec("hash-object", "-t", "tree", "/dev/null")
			if treeErr != nil {
				return fmt.Errorf("creating empty tree: %w", treeErr)
			}
			diffArgs = append(diffArgs, emptyTree)
		} else {
			diffArgs = append(diffArgs, parent)
		}
	}

	diff, err := git.Exec(diffArgs...)
	if err != nil {
		return fmt.Errorf("getting diff: %w", err)
	}

	if diff == "" {
		if amend {
			return fmt.Errorf("no changes found to amend")
		}
		return fmt.Errorf("no staged changes found. Stage your changes with 'git add' first")
	}

	branch, _ := git.Exec("rev-parse", "--abbrev-ref", "HEAD")

	prompt := diff
	if issue := git.ExtractIssue(branch); issue != "" {
		prompt = fmt.Sprintf("Branch: %s (Issue: %s)\n\n%s", branch, issue, diff)
	}

	provider, err := providers.ResolveByAPIKey()
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Fprint(os.Stderr, ".")
			}
		}
	}()

	msg, err := provider.Complete(ctx, systemPrompt, prompt)
	close(done)
	fmt.Fprintln(os.Stderr)

	if err != nil {
		return err
	}

	commitArgs := []string{"commit", "-m", msg, "-e"}
	if amend {
		commitArgs = []string{"commit", "--amend", "-m", msg, "-e"}
	}

	return git.ExecInteractive(commitArgs...)
}
