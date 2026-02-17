package commit

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"regexp"
	"time"

	"llm/internal/git"
	"llm/internal/providers"
)

//go:embed prompt.md
var systemPrompt string

type Config struct {
	Amend bool
}

func ParseConfig(args []string) (*Config, []string, error) {
	cfg := &Config{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-a", "--amend":
			cfg.Amend = true
		default:
			remaining = append(remaining, args[i])
		}
	}

	return cfg, remaining, nil
}

func BuildPrompt(diff, branch string) string {
	if diff == "" {
		return ""
	}

	if issue := extractIssue(branch); issue != "" {
		return fmt.Sprintf("Branch: %s (Issue: %s)\n\n%s", branch, issue, diff)
	}

	return diff
}

func GenerateCommitMessage(ctx context.Context, provider providers.Provider, prompt string, stderr io.Writer, tickerFunc func(time.Duration) *time.Ticker) (string, error) {
	if stderr == nil {
		stderr = io.Discard
	}
	if tickerFunc == nil {
		tickerFunc = func(d time.Duration) *time.Ticker { return time.NewTicker(d) }
	}

	done := make(chan struct{})
	go func() {
		ticker := tickerFunc(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_, _ = fmt.Fprint(stderr, ".")
			}
		}
	}()

	msg, err := provider.Complete(ctx, systemPrompt, prompt)
	close(done)
	_, _ = fmt.Fprintln(stderr)

	if err != nil {
		return "", err
	}

	return msg, nil
}

func Run(ctx context.Context, provider providers.Provider, git git.Client, stderr io.Writer, args []string) error {
	cfg, _, err := ParseConfig(args)
	if err != nil {
		return err
	}

	diff, err := getDiffForCommit(git, cfg.Amend)
	if err != nil {
		return err
	}

	if diff == "" {
		if cfg.Amend {
			return fmt.Errorf("no changes found to amend")
		}
		return fmt.Errorf("no staged changes found. Stage your changes with 'git add' first")
	}

	branch, _ := git.GetCurrentBranch()
	prompt := BuildPrompt(diff, branch)

	msg, err := GenerateCommitMessage(ctx, provider, prompt, stderr, nil)
	if err != nil {
		return err
	}

	return git.Commit(msg, cfg.Amend)
}

func extractIssue(branch string) string {
	re := regexp.MustCompile(`[A-Z]+-\d+`)
	return re.FindString(branch)
}

func getDiffForCommit(git git.Client, amend bool) (string, error) {
	if !amend {
		return git.GetStagedDiff()
	}

	hasParent, _ := git.HasParentCommit()
	if !hasParent {
		return git.GetStagedDiff()
	}

	return git.GetDiffFromRevision("HEAD~1")
}
