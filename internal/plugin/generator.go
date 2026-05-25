// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-md Authors

package plugin

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	breakingChangesSection = "Breaking Changes"
	featuresSection        = "Features"
	bugFixesSection        = "Bug Fixes"
	otherChangesSection    = "Other Changes"
)

type ReleaseContext struct {
	Version        string
	CurrentVersion string
	Branch         string
	Commits        []string
}

type Generator struct {
	now func() time.Time
}

var conventionalHeaderPattern = regexp.MustCompile(`^(\w+)(\([\w-]+\))?(!)?:(.+)$`)

func New() *Generator {
	return &Generator{now: time.Now}
}

func (g *Generator) Generate(ctx ReleaseContext) string {
	sections := map[string][]string{}
	for _, commit := range ctx.Commits {
		section, line := classifyCommit(commit)
		if section == "" || line == "" {
			continue
		}
		sections[section] = append(sections[section], line)
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "## %s (%s)", displayVersion(ctx.Version), g.currentDate().Format("2006-01-02"))

	for _, section := range []string{breakingChangesSection, featuresSection, bugFixesSection, otherChangesSection} {
		lines := sections[section]
		if len(lines) == 0 {
			continue
		}

		builder.WriteString("\n\n### ")
		builder.WriteString(section)
		builder.WriteString("\n")
		for index, line := range lines {
			if index > 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString("- ")
			builder.WriteString(line)
		}
	}

	return builder.String()
}

func (g *Generator) currentDate() time.Time {
	if g != nil && g.now != nil {
		return g.now()
	}
	return time.Now()
}

func classifyCommit(commit string) (string, string) {
	trimmed := strings.TrimSpace(commit)
	if trimmed == "" {
		return "", ""
	}

	if breaking, ok := breakingChangeText(trimmed); ok {
		return breakingChangesSection, breaking
	}

	header := firstLine(trimmed)
	matches := conventionalHeaderPattern.FindStringSubmatch(header)
	if len(matches) == 0 {
		return otherChangesSection, header
	}

	if matches[3] == "!" {
		return breakingChangesSection, header
	}

	switch strings.ToLower(matches[1]) {
	case "feat":
		return featuresSection, header
	case "fix", "perf", "revert":
		return bugFixesSection, header
	default:
		return otherChangesSection, header
	}
}

func breakingChangeText(message string) (string, bool) {
	for _, line := range strings.Split(message, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BREAKING CHANGE:") {
			return trimmed, true
		}
	}
	return "", false
}

func firstLine(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}
	parts := strings.SplitN(message, "\n", 2)
	return strings.TrimSpace(parts[0])
}

func displayVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "Unreleased"
	}
	if strings.HasPrefix(strings.ToLower(version), "v") {
		return version
	}
	return "v" + version
}
