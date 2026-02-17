package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client interface {
	GetStagedDiff() (string, error)
	GetDiffFromRevision(revision string) (string, error)
	GetCurrentBranch() (string, error)
	HasParentCommit() (bool, error)
	Commit(msg string, amend bool) error
}

type RealClient struct{}

func (r *RealClient) GetStagedDiff() (string, error) {
	diff, err := r.exec("diff", "--staged")
	if err != nil {
		return "", fmt.Errorf("getting staged diff: %w", err)
	}
	return diff, nil
}

func (r *RealClient) GetDiffFromRevision(revision string) (string, error) {
	diff, err := r.exec("diff", "--staged", revision)
	if err != nil {
		return "", fmt.Errorf("getting diff from %s: %w", revision, err)
	}
	return diff, nil
}

func (r *RealClient) GetCurrentBranch() (string, error) {
	return r.exec("branch", "--show-current")
}

func (r *RealClient) HasParentCommit() (bool, error) {
	_, err := r.exec("rev-parse", "--verify", "HEAD~1")
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (r *RealClient) Commit(msg string, amend bool) error {
	commitArgs := []string{"commit", "-m", msg, "-e"}
	if amend {
		commitArgs = []string{"commit", "--amend", "-m", msg, "-e"}
	}

	return r.execInteractive(commitArgs...)
}

func (r *RealClient) exec(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (r *RealClient) execInteractive(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
