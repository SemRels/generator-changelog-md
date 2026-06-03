// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-md Authors

package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGeneratorGenerate(t *testing.T) {
	t.Parallel()

	generator := &Generator{now: func() time.Time {
		return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	}}

	output := generator.Generate(ReleaseContext{
		Version:       "1.3.0",
		RepositoryURL: "https://github.com/SemRels/semrel",
		Commits: []string{
			"feat: add new feature (#123)",
			"fix: resolve issue with X (#124)",
			"perf: improve cache warmup (#125)",
			"refactor!: remove deprecated API",
			"docs: update README",
			"feat: adjust auth\n\nBREAKING CHANGE: API changed",
		},
	})

	require.Equal(t, "## v1.3.0 (2026-05-25)\n\n### Breaking Changes\n- refactor!: remove deprecated API\n- BREAKING CHANGE: API changed\n\n### Features\n- feat: add new feature ([#123](https://github.com/SemRels/semrel/pull/123))\n\n### Performance Improvements\n- perf: improve cache warmup ([#125](https://github.com/SemRels/semrel/pull/125))\n\n### Bug Fixes\n- fix: resolve issue with X ([#124](https://github.com/SemRels/semrel/pull/124))\n\n### Other Changes\n- docs: update README", output)
}

func TestGeneratorGenerateWithoutVersionUsesUnreleased(t *testing.T) {
	t.Parallel()

	generator := &Generator{now: func() time.Time {
		return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	}}

	output := generator.Generate(ReleaseContext{})

	require.Equal(t, "## Unreleased (2026-05-25)", output)
}

func TestGeneratorGenerateFlatMode(t *testing.T) {
	t.Parallel()

	generator := &Generator{now: func() time.Time {
		return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	}}

	output := generator.Generate(ReleaseContext{
		Version:       "1.3.0",
		RepositoryURL: "https://github.com/SemRels/semrel",
		Commits: []string{
			"feat: add new feature (#123)",
			"perf: improve cache warmup",
			"docs: update README",
		},
	}, GenerateOptions{GroupByType: false, LinkPRs: true, LinkCommits: true})

	require.Equal(t, "## v1.3.0 (2026-05-25)\n- feat: add new feature ([#123](https://github.com/SemRels/semrel/pull/123))\n- perf: improve cache warmup\n- docs: update README", output)
}

func TestLinkifyCommitTextWithoutRepositoryURL(t *testing.T) {
	t.Parallel()

	line := linkifyCommitText("feat: add new feature (#123)", ReleaseContext{}, DefaultGenerateOptions())

	require.Equal(t, "feat: add new feature (#123)", line)
}

func TestClassifyCommit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		commit      string
		wantSection string
		wantLine    string
	}{
		{
			name:        "empty",
			commit:      "  ",
			wantSection: "",
			wantLine:    "",
		},
		{
			name:        "feature",
			commit:      "feat(api): add login",
			wantSection: featuresSection,
			wantLine:    "feat(api): add login",
		},
		{
			name:        "performance",
			commit:      "perf: improve query speed",
			wantSection: performanceSection,
			wantLine:    "perf: improve query speed",
		},
		{
			name:        "breaking footer",
			commit:      "feat: add login\n\nBREAKING CHANGE: API changed",
			wantSection: breakingChangesSection,
			wantLine:    "BREAKING CHANGE: API changed",
		},
		{
			name:        "other",
			commit:      "update generated files",
			wantSection: otherChangesSection,
			wantLine:    "update generated files",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			section, line := classifyCommit(tt.commit, ReleaseContext{}, DefaultGenerateOptions())
			require.Equal(t, tt.wantSection, section)
			require.Equal(t, tt.wantLine, line)
		})
	}
}
