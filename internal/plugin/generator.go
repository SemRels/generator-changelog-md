// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-semrel Authors

// Package plugin implements the markdown changelog generator business logic.
package plugin

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	semrelv1 "github.com/SemRels/generator-changelog-md/internal/gen/v1"
)

// commitGroup holds a heading and the list of formatted commit lines beneath it.
type commitGroup struct {
	heading string
	lines   []string
}

// ParsedCommit is the result of parsing a single conventional-commit message.
type ParsedCommit struct {
	Type        string
	Scope       string
	Breaking    bool
	Description string
	SHA         string
}

// conventionalCommitRE matches:  type(scope)!: description  OR  type!: desc  OR  type: desc
var conventionalCommitRE = regexp.MustCompile(
	`^(?P<type>[a-zA-Z]+)(?:\((?P<scope>[^)]*)\))?(?P<bang>!)?:\s*(?P<desc>.+)$`,
)

// ParseCommit parses a conventional-commit raw_message.
// The first line is used for type/scope/description detection.
// The full message is scanned for a "BREAKING CHANGE:" footer.
func ParseCommit(sha, rawMessage string) ParsedCommit {
	lines := strings.SplitN(strings.TrimSpace(rawMessage), "\n", 2)
	firstLine := strings.TrimSpace(lines[0])

	pc := ParsedCommit{SHA: sha, Description: firstLine}

	// Scan entire message for BREAKING CHANGE footer.
	if strings.Contains(rawMessage, "BREAKING CHANGE:") ||
		strings.Contains(rawMessage, "BREAKING-CHANGE:") {
		pc.Breaking = true
	}

	m := conventionalCommitRE.FindStringSubmatch(firstLine)
	if m == nil {
		// Not a conventional commit – keep raw description.
		return pc
	}

	names := conventionalCommitRE.SubexpNames()
	for i, name := range names {
		switch name {
		case "type":
			pc.Type = strings.ToLower(m[i])
		case "scope":
			pc.Scope = m[i]
		case "bang":
			if m[i] == "!" {
				pc.Breaking = true
			}
		case "desc":
			pc.Description = m[i]
		}
	}

	return pc
}

// GroupedCommits holds the four sections shown in the changelog.
type GroupedCommits struct {
	BreakingChanges []ParsedCommit
	Features        []ParsedCommit
	Fixes           []ParsedCommit
	Others          []ParsedCommit
}

// GroupCommits classifies a slice of proto Commits into the four sections.
func GroupCommits(commits []*semrelv1.Commit) GroupedCommits {
	var g GroupedCommits

	for _, c := range commits {
		if c == nil {
			continue
		}
		pc := ParseCommit(c.GetSha(), c.GetRawMessage())

		if pc.Breaking {
			g.BreakingChanges = append(g.BreakingChanges, pc)
			continue
		}
		switch pc.Type {
		case "feat", "feature":
			g.Features = append(g.Features, pc)
		case "fix", "bugfix":
			g.Fixes = append(g.Fixes, pc)
		default:
			g.Others = append(g.Others, pc)
		}
	}

	return g
}

// versionString converts a SemanticVersion into the canonical "vMAJOR.MINOR.PATCH[-pre]" form.
func versionString(v *semrelv1.SemanticVersion) string {
	if v == nil {
		return "v0.0.0"
	}
	s := fmt.Sprintf("v%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
	if pre := v.GetPreRelease(); pre != "" {
		s += "-" + pre
	}
	return s
}

// formatLine produces a single bullet line for a parsed commit.
func formatLine(pc ParsedCommit) string {
	if pc.Scope != "" {
		return fmt.Sprintf("- **%s**: %s", pc.Scope, pc.Description)
	}
	return "- " + pc.Description
}

// FormatMarkdown renders a complete changelog section for the release described
// by ctx.  It returns the markdown string (without trailing newline).
func FormatMarkdown(ctx *semrelv1.ReleaseContext) string {
	if ctx == nil {
		return ""
	}

	version := versionString(ctx.GetNextVersion())
	date := time.Now().UTC().Format("2006-01-02")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## [%s] - %s\n", version, date))

	grouped := GroupCommits(ctx.GetCommits())

	groups := []commitGroup{
		{heading: "### ⚠ Breaking Changes", lines: commitLines(grouped.BreakingChanges)},
		{heading: "### Features", lines: commitLines(grouped.Features)},
		{heading: "### Fixes", lines: commitLines(grouped.Fixes)},
		{heading: "### Other Changes", lines: commitLines(grouped.Others)},
	}

	for _, g := range groups {
		if len(g.lines) == 0 {
			continue
		}
		sb.WriteString("\n")
		sb.WriteString(g.heading)
		sb.WriteString("\n\n")
		for _, l := range g.lines {
			sb.WriteString(l)
			sb.WriteString("\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func commitLines(pcs []ParsedCommit) []string {
	if len(pcs) == 0 {
		return nil
	}
	out := make([]string, len(pcs))
	for i, pc := range pcs {
		out[i] = formatLine(pc)
	}
	return out
}

// PrependToChangelog inserts newSection at the top of existing, preserving any
// existing "# Changelog" / "# CHANGELOG" title if present.
func PrependToChangelog(newSection, existing string) string {
	existing = strings.TrimSpace(existing)

	if existing == "" {
		return "# Changelog\n\n" + newSection
	}

	lines := strings.SplitN(existing, "\n", 2)
	firstLine := strings.TrimSpace(lines[0])

	if strings.HasPrefix(firstLine, "# ") {
		// Preserve the title and insert after it.
		rest := ""
		if len(lines) > 1 {
			rest = strings.TrimLeft(lines[1], "\n")
		}
		if rest == "" {
			return existing + "\n\n" + newSection
		}
		return firstLine + "\n\n" + newSection + "\n\n" + rest
	}

	// No title – just prepend.
	return newSection + "\n\n" + existing
}

// AppendToChangelog adds newSection at the bottom of existing.
func AppendToChangelog(newSection, existing string) string {
	existing = strings.TrimSpace(existing)
	if existing == "" {
		return "# Changelog\n\n" + newSection
	}
	return existing + "\n\n" + newSection
}

// WriteChangelog writes the full changelog content to the given file path,
// optionally prepending the new section to any existing content.
func WriteChangelog(path, newSection string, prepend bool) error {
	existing := ""
	data, err := os.ReadFile(path)
	if err == nil {
		existing = string(data)
	}

	var content string
	if prepend {
		content = PrependToChangelog(newSection, existing)
	} else {
		content = AppendToChangelog(newSection, existing)
	}

	return os.WriteFile(path, []byte(content+"\n"), 0o644)
}
