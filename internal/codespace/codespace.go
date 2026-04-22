// Package codespace wraps the GitHub CLI (gh) to list, create, delete,
// and SSH into GitHub Codespaces. All GitHub API interactions go through
// the GHRunner interface so they can be stubbed in tests.
package codespace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ananth/cosmonaut/internal/config"
)

// RepoField handles the gh CLI returning repository as either a string or an object.
type RepoField string

func (r *RepoField) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*r = RepoField(s)
		return nil
	}

	var obj struct {
		FullName      string `json:"full_name"`
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("repository field is neither string nor object: %w", err)
	}
	name := obj.FullName
	if name == "" {
		name = obj.NameWithOwner
	}
	*r = RepoField(name)
	return nil
}

type GitStatus struct {
	Ref    string `json:"ref,omitempty"`
	Branch string `json:"branch,omitempty"`
}

type Codespace struct {
	Name        string     `json:"name"`
	DisplayName string     `json:"displayName,omitempty"`
	Repository  RepoField  `json:"repository"`
	State       string     `json:"state,omitempty"`
	GitStatus   *GitStatus `json:"gitStatus,omitempty"`
}

// GHRunner abstracts the GitHub CLI for testability.
type GHRunner interface {
	Run(args []string) (string, error)
	RunMerged(args []string) (string, error)
	// RunInteractive runs a command with stdin connected and output teed to
	// both a capture buffer and the real terminal. Use this when the gh
	// subprocess may need to prompt the user (e.g. machine-type selection).
	RunInteractive(args []string) (string, error)
}

// DefaultGHRunner executes real gh commands.
type DefaultGHRunner struct{}

func (d DefaultGHRunner) Run(args []string) (string, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("gh %s exited with code %d: %s", strings.Join(args, " "), cmd.ProcessState.ExitCode(), detail)
	}
	return stdout.String(), nil
}

func (d DefaultGHRunner) RunMerged(args []string) (string, error) {
	cmd := exec.Command("gh", args...)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return combined.String(), fmt.Errorf("gh %s exited with code %d: %s", strings.Join(args, " "), cmd.ProcessState.ExitCode(), combined.String())
	}
	return combined.String(), nil
}

func (d DefaultGHRunner) RunInteractive(args []string) (string, error) {
	cmd := exec.Command("gh", args...)
	var buf bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(&buf, os.Stderr)
	cmd.Stderr = io.MultiWriter(&buf, os.Stderr)
	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("gh %s exited with code %d: %s", strings.Join(args, " "), cmd.ProcessState.ExitCode(), buf.String())
	}
	return buf.String(), nil
}

// RequireCommand checks that a command is on PATH.
func RequireCommand(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%q not found on PATH", name)
	}
	return nil
}

// EnsureGHAuth verifies the user is authenticated with gh.
func EnsureGHAuth(runner GHRunner) error {
	_, err := runner.Run([]string{"auth", "status"})
	if err != nil {
		return fmt.Errorf("GitHub CLI is not authenticated. Run `gh auth login` first")
	}
	return nil
}

// ListCodespaces returns all codespaces for a given repository.
func ListCodespaces(runner GHRunner, repository string) ([]Codespace, error) {
	out, err := runner.Run([]string{
		"codespace", "list",
		"--repo", repository,
		"--json", "name,displayName,repository,state,gitStatus",
	})
	if err != nil {
		return nil, err
	}

	var codespaces []Codespace
	if err := json.Unmarshal([]byte(out), &codespaces); err != nil {
		return nil, fmt.Errorf("parsing codespace list: %w", err)
	}
	return codespaces, nil
}

// ListAllCodespaces returns all codespaces across all repos.
func ListAllCodespaces(runner GHRunner) ([]Codespace, error) {
	out, err := runner.Run([]string{
		"codespace", "list",
		"--json", "name,displayName,repository,state,gitStatus",
	})
	if err != nil {
		return nil, err
	}

	var codespaces []Codespace
	if err := json.Unmarshal([]byte(out), &codespaces); err != nil {
		return nil, fmt.Errorf("parsing codespace list: %w", err)
	}
	return codespaces, nil
}

// UniqueRepos extracts sorted unique repository names from codespaces.
func UniqueRepos(codespaces []Codespace) []string {
	seen := make(map[string]bool)
	var repos []string
	for _, cs := range codespaces {
		repo := string(cs.Repository)
		if repo != "" && !seen[repo] {
			seen[repo] = true
			repos = append(repos, repo)
		}
	}
	return repos
}

// ListAllRepos returns all repositories the user has access to.
func ListAllRepos(runner GHRunner) ([]string, error) {
	out, err := runner.Run([]string{
		"repo", "list",
		"--json", "nameWithOwner",
		"--limit", "1000",
	})
	if err != nil {
		return nil, err
	}

	var repos []struct {
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := json.Unmarshal([]byte(out), &repos); err != nil {
		return nil, fmt.Errorf("parsing repo list: %w", err)
	}

	result := make([]string, len(repos))
	for i, r := range repos {
		result[i] = r.NameWithOwner
	}
	return result, nil
}

// FilterByRepo returns codespaces belonging to a given repository.
func FilterByRepo(codespaces []Codespace, repo string) []Codespace {
	var result []Codespace
	for _, cs := range codespaces {
		if string(cs.Repository) == repo {
			result = append(result, cs)
		}
	}
	return result
}

var codespaceNameRe = regexp.MustCompile(`[A-Za-z0-9-]+-[A-Za-z0-9]{6,}`)

// CreateCodespace creates a new codespace and returns its details.
func CreateCodespace(runner GHRunner, target config.Target) (*Codespace, error) {
	args := BuildCreateArgs(target)
	combined, err := runner.RunMerged(args[1:]) // strip "gh" prefix
	if err != nil && combined == "" {
		return nil, err
	}

	match := codespaceNameRe.FindString(combined)
	if match == "" {
		return nil, fmt.Errorf("codespace created but name could not be determined from gh output:\n%s", strings.TrimSpace(combined))
	}

	out, err := runner.Run([]string{
		"codespace", "view",
		"--codespace", match,
		"--json", "name,displayName,repository,state,gitStatus",
	})
	if err != nil {
		return nil, err
	}

	var cs Codespace
	if err := json.Unmarshal([]byte(out), &cs); err != nil {
		return nil, fmt.Errorf("parsing codespace view: %w", err)
	}
	return &cs, nil
}

// CreateCodespaceInteractive creates a codespace with terminal access so gh can
// prompt the user if needed (e.g. machine type selection). Output is shown on
// the terminal in real time.
func CreateCodespaceInteractive(runner GHRunner, target config.Target) (*Codespace, error) {
	args := BuildCreateArgs(target)
	combined, err := runner.RunInteractive(args[1:]) // strip "gh" prefix
	if err != nil && combined == "" {
		return nil, err
	}

	match := codespaceNameRe.FindString(combined)
	if match == "" {
		return nil, fmt.Errorf("codespace created but name could not be determined from gh output:\n%s", strings.TrimSpace(combined))
	}

	out, err := runner.Run([]string{
		"codespace", "view",
		"--codespace", match,
		"--json", "name,displayName,repository,state,gitStatus",
	})
	if err != nil {
		return nil, err
	}

	var cs Codespace
	if err := json.Unmarshal([]byte(out), &cs); err != nil {
		return nil, fmt.Errorf("parsing codespace view: %w", err)
	}
	return &cs, nil
}

// DeleteCodespace deletes a codespace by name.
func DeleteCodespace(runner GHRunner, name string) error {
	_, err := runner.Run([]string{"codespace", "delete", "--codespace", name, "--force"})
	return err
}

// EnsureReachable verifies the codespace SSH server is accessible.
func EnsureReachable(runner GHRunner, codespaceName string) error {
	_, err := runner.Run([]string{"codespace", "ssh", "--codespace", codespaceName, "--", "true"})
	if err != nil {
		return fmt.Errorf("could not start or SSH into codespace %q: %w", codespaceName, err)
	}
	return nil
}

// GetSSHConfig retrieves the OpenSSH config for a codespace.
func GetSSHConfig(runner GHRunner, codespaceName string) (string, error) {
	return runner.Run([]string{"codespace", "ssh", "--codespace", codespaceName, "--config"})
}
