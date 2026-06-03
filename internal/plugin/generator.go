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
	performanceSection     = "Performance Improvements"
	bugFixesSection        = "Bug Fixes"
	otherChangesSection    = "Other Changes"
)

type ReleaseContext struct {
	Version        string
	CurrentVersion string
	Branch         string
	RepositoryURL  string
	Commits        []string
}

type GenerateOptions struct {
	GroupByType bool
	LinkPRs     bool
	LinkCommits bool
}

type Generator struct {
	now func() time.Time
}

var (
	conventionalHeaderPattern = regexp.MustCompile(`^(\w+)(\([\w-]+\))?(!)?:(.+)$`)
	pullRequestPattern        = regexp.MustCompile(`\(#(\d+)\)`)
	barePullRequestPattern    = regexp.MustCompile(`(^|[^\[\w])#(\d+)`)
	commitSHAPattern          = regexp.MustCompile(`\b[0-9a-f]{40}\b`)
)

func New() *Generator {
	return &Generator{now: time.Now}
}

func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		GroupByType: true,
		LinkPRs:     true,
		LinkCommits: true,
	}
}

func (g *Generator) Generate(ctx ReleaseContext, options ...GenerateOptions) string {
	generateOptions := DefaultGenerateOptions()
	if len(options) > 0 {
		generateOptions = options[0]
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "## %s (%s)", displayVersion(ctx.Version), g.currentDate().Format("2006-01-02"))

	if !generateOptions.GroupByType {
		lines := g.commitLines(ctx, generateOptions)
		if len(lines) == 0 {
			return builder.String()
		}
		builder.WriteString("\n")
		writeLines(&builder, lines)
		return builder.String()
	}

	sections := map[string][]string{}
	for _, commit := range ctx.Commits {
		section, line := classifyCommit(commit, ctx, generateOptions)
		if section == "" || line == "" {
			continue
		}
		sections[section] = append(sections[section], line)
	}

	for _, section := range []string{breakingChangesSection, featuresSection, performanceSection, bugFixesSection, otherChangesSection} {
		lines := sections[section]
		if len(lines) == 0 {
			continue
		}

		builder.WriteString("\n\n### ")
		builder.WriteString(section)
		builder.WriteString("\n")
		writeLines(&builder, lines)
	}

	return builder.String()
}

func (g *Generator) commitLines(ctx ReleaseContext, options GenerateOptions) []string {
	lines := make([]string, 0, len(ctx.Commits))
	for _, commit := range ctx.Commits {
		_, line := classifyCommit(commit, ctx, options)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func writeLines(builder *strings.Builder, lines []string) {
	for index, line := range lines {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString("- ")
		builder.WriteString(line)
	}
}

func (g *Generator) currentDate() time.Time {
	if g != nil && g.now != nil {
		return g.now()
	}
	return time.Now()
}

func classifyCommit(commit string, ctx ReleaseContext, options GenerateOptions) (string, string) {
	trimmed := strings.TrimSpace(commit)
	if trimmed == "" {
		return "", ""
	}

	if breaking, ok := breakingChangeText(trimmed); ok {
		return breakingChangesSection, linkifyCommitText(breaking, ctx, options)
	}

	header := linkifyCommitText(firstLine(trimmed), ctx, options)
	matches := conventionalHeaderPattern.FindStringSubmatch(firstLine(trimmed))
	if len(matches) == 0 {
		return otherChangesSection, header
	}

	if matches[3] == "!" {
		return breakingChangesSection, header
	}

	switch strings.ToLower(matches[1]) {
	case "feat":
		return featuresSection, header
	case "perf":
		return performanceSection, header
	case "fix", "revert":
		return bugFixesSection, header
	default:
		return otherChangesSection, header
	}
}

func linkifyCommitText(text string, ctx ReleaseContext, options GenerateOptions) string {
	if text == "" {
		return text
	}

	repositoryURL := strings.TrimRight(strings.TrimSpace(ctx.RepositoryURL), "/")
	if repositoryURL == "" {
		return text
	}

	if options.LinkPRs {
		text = pullRequestPattern.ReplaceAllString(text, `([#$1](`+repositoryURL+`/pull/$1))`)
		text = barePullRequestPattern.ReplaceAllString(text, `${1}[#$2](`+repositoryURL+`/pull/$2)`)
	}

	if options.LinkCommits {
		text = commitSHAPattern.ReplaceAllStringFunc(text, func(sha string) string {
			return fmt.Sprintf("[%s](%s/commit/%s)", sha, repositoryURL, sha)
		})
	}

	return text
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
