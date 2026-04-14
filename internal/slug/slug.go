// Package slug converts free-form work labels into URL-friendly slugs
// and composes codespace display names from repo, branch, and label parts.
package slug

import (
	"regexp"
	"strings"
)

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

// SlugifyWorkLabel converts a free-form label into a URL-friendly slug.
func SlugifyWorkLabel(value string) string {
	lowered := strings.ToLower(strings.TrimSpace(value))
	slug := nonAlphanumRe.ReplaceAllString(lowered, "-")
	return strings.Trim(slug, "-")
}

// BuildDisplayName composes a codespace display name from repo, branch, and work label.
// The result is truncated to 48 characters (GitHub Codespaces limit).
func BuildDisplayName(repository, branch, workLabel, fallback string) string {
	parts := strings.SplitN(repository, "/", 2)
	repoName := parts[len(parts)-1]

	baseParts := []string{repoName}
	if branch != "" {
		baseParts = append(baseParts, branch)
	}
	base := strings.Join(baseParts, "-")

	s := SlugifyWorkLabel(workLabel)
	var candidate string
	if s != "" {
		candidate = base + "-" + s
	} else if fallback != "" {
		candidate = fallback
	} else {
		candidate = base
	}

	if len(candidate) > 48 {
		candidate = candidate[:48]
	}
	return candidate
}
