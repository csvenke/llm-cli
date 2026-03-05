package gh

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var lookPathExec = exec.LookPath
var execCommand = exec.Command

type Client interface {
	GetCurrentBranch() (string, error)
	GetDefaultBranch() (string, error)
	GetMergeBase(base, head string) (string, error)
	GetDiffRange(from, to string) (string, error)
}

type RealClient struct{}

func (r *RealClient) GetCurrentBranch() (string, error) {
	return r.exec("branch", "--show-current")
}

func (r *RealClient) GetDefaultBranch() (string, error) {
	return r.exec("rev-parse", "--abbrev-ref", "--symbolic-full-name", "refs/remotes/origin/HEAD")
}

func (r *RealClient) GetMergeBase(base, head string) (string, error) {
	return r.exec("merge-base", base, head)
}

func (r *RealClient) GetDiffRange(from, to string) (string, error) {
	return r.exec("diff", fmt.Sprintf("%s..%s", from, to))
}

func (r *RealClient) exec(args ...string) (string, error) {
	cmd := execCommand("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}
