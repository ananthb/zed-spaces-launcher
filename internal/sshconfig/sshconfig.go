// Package sshconfig manages the local SSH configuration for codespace
// connections. It writes per-codespace config files into
// ~/.ssh/codespaces-zed/ and ensures the main ~/.ssh/config includes them.
package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const SSHIncludeLine = "Include ~/.ssh/codespaces-zed/*.conf"

var hostAliasRe = regexp.MustCompile(`(?m)^\s*Host\s+([^\s*][^\s]*)\s*$`)

// ParsePrimaryHostAlias extracts the first concrete Host entry from SSH config text.
func ParsePrimaryHostAlias(sshConfig string) (string, error) {
	match := hostAliasRe.FindStringSubmatch(sshConfig)
	if match == nil {
		return "", fmt.Errorf("could not find a concrete Host entry in gh codespace ssh --config output")
	}
	return match[1], nil
}

// EnsureIncludeLine ensures the SSH include line is at the top of the config.
// It removes any existing copy and prepends it.
func EnsureIncludeLine(configText string) string {
	var lines []string
	for _, line := range strings.Split(configText, "\n") {
		if strings.TrimSpace(line) != SSHIncludeLine {
			lines = append(lines, line)
		}
	}
	body := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	if body != "" {
		return SSHIncludeLine + "\n" + body + "\n"
	}
	return SSHIncludeLine + "\n"
}

// SSHPaths holds the resolved SSH directory paths.
type SSHPaths struct {
	MainConfigPath string
	IncludeDir     string
}

// ResolvePaths returns the SSH paths for the current platform.
func ResolvePaths() SSHPaths {
	home, _ := os.UserHomeDir()
	sshDir := filepath.Join(home, ".ssh")
	return SSHPaths{
		MainConfigPath: filepath.Join(sshDir, "config"),
		IncludeDir:     filepath.Join(sshDir, "codespaces-zed"),
	}
}

// CodespaceConfigPath returns the path for a codespace-specific SSH config.
func (p SSHPaths) CodespaceConfigPath(codespaceName string) string {
	return filepath.Join(p.IncludeDir, codespaceName+".conf")
}

// EnsureConfigIncludesGenerated ensures the main SSH config includes the generated configs.
func EnsureConfigIncludesGenerated(mainConfigPath string) error {
	current, err := os.ReadFile(mainConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	currentStr := string(current)
	if strings.Contains(currentStr, SSHIncludeLine) {
		return nil
	}

	updated := EnsureIncludeLine(currentStr)
	dir := filepath.Dir(mainConfigPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(mainConfigPath, []byte(updated), 0644)
}

// ReadExistingAlias reads the SSH alias from an existing codespace config file.
// Returns the alias and true if the file exists and contains a valid Host entry,
// or empty string and false otherwise.
func ReadExistingAlias(includeDir, codespaceName string) (string, bool) {
	path := filepath.Join(includeDir, codespaceName+".conf")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	alias, err := ParsePrimaryHostAlias(string(data))
	if err != nil {
		return "", false
	}
	return alias, true
}

// WriteCodespaceConfig writes the SSH config for a specific codespace.
func WriteCodespaceConfig(includeDir, codespaceName, content string) error {
	if err := os.MkdirAll(includeDir, 0700); err != nil {
		return err
	}
	path := filepath.Join(includeDir, codespaceName+".conf")
	return os.WriteFile(path, []byte(content), 0644)
}
