// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The SemRels Authors

package plugin

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	semrelv1 "github.com/SemRels/semrel-api/api/gen/v1"
)

const (
	featuresSection      = "✨ Features"
	bugFixesSection      = "🐛 Bug Fixes"
	performanceSection   = "⚡ Performance"
	documentationSection = "📚 Documentation"
	otherChangesSection  = "🔄 Other Changes"
)

var (
	conventionalCommitPattern = regexp.MustCompile(`^(\w+)(\([\w-]+\))?(!)?:(.+)$`)
	nowFunc                   = time.Now
	sectionOrder              = []string{
		featuresSection,
		bugFixesSection,
		performanceSection,
		documentationSection,
		otherChangesSection,
	}
)

type Generator struct {
	semrelv1.UnimplementedChangelogGeneratorPluginServer
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateNotes(_ context.Context, req *semrelv1.GenerateNotesRequest) (*semrelv1.GenerateNotesResponse, error) {
	if req == nil {
		return &semrelv1.GenerateNotesResponse{}, nil
	}

	return &semrelv1.GenerateNotesResponse{Notes: renderNotes(req.GetCtx())}, nil
}

func renderNotes(ctx *semrelv1.ReleaseContext) string {
	var builder strings.Builder
	sections := map[string][]string{}

	if ctx != nil {
		for _, commit := range ctx.GetCommits() {
			section, line := renderCommit(commit)
			sections[section] = append(sections[section], line)
		}

		if nextVersion := ctx.GetNextVersion(); nextVersion != nil {
			fmt.Fprintf(&builder, "## v%d.%d.%d (%s)", nextVersion.GetMajor(), nextVersion.GetMinor(), nextVersion.GetPatch(), nowFunc().Format("2006-01-02"))
		}
	}

	for _, section := range sectionOrder {
		lines := sections[section]
		if len(lines) == 0 {
			continue
		}

		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}

		fmt.Fprintf(&builder, "### %s\n", section)
		for index, line := range lines {
			if index > 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString(line)
		}
	}

	return builder.String()
}

func renderCommit(commit *semrelv1.Commit) (string, string) {
	rawMessage := ""
	sha := ""
	if commit != nil {
		rawMessage = commit.GetRawMessage()
		sha = commit.GetSha()
	}

	header := firstLine(rawMessage)
	matches := conventionalCommitPattern.FindStringSubmatch(header)
	breaking := strings.Contains(rawMessage, "BREAKING CHANGE:")
	section := otherChangesSection
	lineText := header

	if len(matches) > 0 {
		commitType := matches[1]
		scope := matches[2]
		if matches[3] == "!" {
			breaking = true
		}

		section = sectionForType(commitType)
		lineText = fmt.Sprintf("%s%s: %s", commitType, scope, strings.TrimSpace(matches[4]))
	}

	if breaking {
		lineText = "💥 **BREAKING CHANGE**: " + lineText
	}

	return section, fmt.Sprintf("- %s (%s)", lineText, shortSHA(sha))
}

func sectionForType(commitType string) string {
	switch commitType {
	case "feat":
		return featuresSection
	case "fix", "revert":
		return bugFixesSection
	case "perf":
		return performanceSection
	case "docs":
		return documentationSection
	default:
		return otherChangesSection
	}
}

func firstLine(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}

	parts := strings.SplitN(message, "\n", 2)
	return strings.TrimSpace(parts[0])
}

func shortSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}

	return sha[:7]
}
