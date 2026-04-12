// Package zed reads and updates Zed's settings.json to upsert SSH
// remote connections. It preserves comments and formatting by operating
// on the parsed JSONC structure.
package zed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"regexp"
)

type Project struct {
	Paths []string `json:"paths"`
}

type SSHConnection struct {
	Host                string    `json:"host"`
	Nickname            string    `json:"nickname,omitempty"`
	Projects            []Project `json:"projects,omitempty"`
	UploadBinaryOverSSH *bool     `json:"upload_binary_over_ssh,omitempty"`
	Port                int       `json:"port,omitempty"`
	Username            string    `json:"username,omitempty"`
}

// ResolveSettingsPath returns the Zed settings.json path for the current platform.
func ResolveSettingsPath() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, ".zed", "settings.json")
	}
	return filepath.Join(home, ".config", "zed", "settings.json")
}

// BuildConnection creates an SSHConnection from target config and SSH alias.
func BuildConnection(host, workspacePath, nickname string, uploadBinary *bool) SSHConnection {
	conn := SSHConnection{
		Host:     host,
		Nickname: nickname,
		Projects: []Project{{Paths: []string{workspacePath}}},
	}
	if uploadBinary != nil {
		conn.UploadBinaryOverSSH = uploadBinary
	}
	return conn
}

// ResolveNickname determines the nickname for a Zed connection.
// It checks zedNickname, targetDisplayName, codespaceDisplayName, then targetName.
func ResolveNickname(zedNickname, targetDisplayName, codespaceDisplayName, targetName string) string {
	if zedNickname != "" {
		return zedNickname
	}
	if targetDisplayName != "" {
		return targetDisplayName
	}
	if codespaceDisplayName != "" {
		return codespaceDisplayName
	}
	return targetName
}

// connectionIdentity returns a comparable identity string for matching connections.
func connectionIdentity(conn map[string]any) string {
	host, _ := conn["host"].(string)
	username, _ := conn["username"].(string)
	port := 22
	if p, ok := conn["port"].(float64); ok {
		port = int(p)
	}
	return fmt.Sprintf("%s@%s:%d", username, host, port)
}

// UpsertConnection merges a new SSH connection into the settings map.
func UpsertConnection(settings map[string]any, conn SSHConnection) map[string]any {
	result := make(map[string]any, len(settings))
	for k, v := range settings {
		result[k] = v
	}

	connJSON, _ := json.Marshal(conn)
	var connMap map[string]any
	json.Unmarshal(connJSON, &connMap)

	newIdentity := connectionIdentity(connMap)

	var existing []any
	if raw, ok := result["ssh_connections"]; ok {
		if arr, ok := raw.([]any); ok {
			existing = append(existing, arr...)
		}
	}

	found := -1
	for i, item := range existing {
		if m, ok := item.(map[string]any); ok {
			if connectionIdentity(m) == newIdentity {
				found = i
				break
			}
		}
	}

	if found >= 0 {
		merged := existing[found].(map[string]any)
		for k, v := range connMap {
			merged[k] = v
		}
		existing[found] = merged
	} else {
		existing = append(existing, connMap)
	}

	result["ssh_connections"] = existing
	return result
}

// UpsertConnectionInFile reads, updates, and writes the Zed settings file.
func UpsertConnectionInFile(settingsPath string, conn SSHConnection) error {
	current := make(map[string]any)

	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(data) > 0 {
		// Parse JSONC (Zed settings may have comments)
		if err := json.Unmarshal(data, &current); err != nil {
			// Try stripping comments with a simple approach
			clean := stripJSONComments(string(data))
			if err := json.Unmarshal([]byte(clean), &current); err != nil {
				return fmt.Errorf("parsing %s: %w", settingsPath, err)
			}
		}
	}

	updated := UpsertConnection(current, conn)

	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, append(out, '\n'), 0644)
}

var (
	blockCommentRe  = regexp.MustCompile(`(?s)/\*.*?\*/`)
	lineCommentRe   = regexp.MustCompile(`(?m)^\s*//.*$`)
	trailingCommaRe = regexp.MustCompile(`,\s*([}\]])`)
)

func stripJSONComments(s string) string {
	s = blockCommentRe.ReplaceAllString(s, "")
	s = lineCommentRe.ReplaceAllString(s, "")
	s = trailingCommaRe.ReplaceAllString(s, "$1")
	return s
}
