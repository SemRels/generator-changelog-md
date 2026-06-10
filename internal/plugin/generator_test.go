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

func TestGeneratorNewContributors(t *testing.T) {
	t.Parallel()

	generator := &Generator{now: func() time.Time {
		return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	}}

	ctx := ReleaseContext{
		Version:       "1.4.0",
		RepositoryURL: "https://github.com/SemRels/semrel",
		Commits:       []string{"feat: add login (#42)"},
	}

	t.Run("new contributors section rendered when contributors provided", func(t *testing.T) {
		t.Parallel()

		opts := DefaultGenerateOptions()
		opts.Contributors = []Contributor{
			{Name: "Alice", Login: "alice", PR: 42},
			{Name: "dependabot[bot]", Login: "dependabot[bot]", PR: 7},
		}

		output := generator.Generate(ctx, opts)

		require.Contains(t, output, "### New Contributors")
		require.Contains(t, output, "[@alice](https://github.com/alice)")
		require.Contains(t, output, "[#42](https://github.com/SemRels/semrel/pull/42)")
		require.Contains(t, output, "[@dependabot[bot]](https://github.com/dependabot[bot])")
		require.Contains(t, output, "[#7](https://github.com/SemRels/semrel/pull/7)")
	})

	t.Run("new contributors section skipped when contributors empty", func(t *testing.T) {
		t.Parallel()

		opts := DefaultGenerateOptions()
		opts.Contributors = nil

		output := generator.Generate(ctx, opts)

		require.NotContains(t, output, "New Contributors")
	})

	t.Run("new contributors section skipped when disabled", func(t *testing.T) {
		t.Parallel()

		opts := DefaultGenerateOptions()
		opts.NewContributors = false
		opts.Contributors = []Contributor{{Name: "Alice", Login: "alice"}}

		output := generator.Generate(ctx, opts)

		require.NotContains(t, output, "New Contributors")
	})

	t.Run("contributor without login uses plain name", func(t *testing.T) {
		t.Parallel()

		opts := DefaultGenerateOptions()
		opts.Contributors = []Contributor{{Name: "Alice Smith"}}

		output := generator.Generate(ctx, opts)

		require.Contains(t, output, "Alice Smith made their first contribution")
	})

	t.Run("contributor without PR omits PR link", func(t *testing.T) {
		t.Parallel()

		opts := DefaultGenerateOptions()
		opts.Contributors = []Contributor{{Name: "Bob", Login: "bob"}}

		ctxNoPR := ReleaseContext{
			Version:       "1.4.0",
			RepositoryURL: "https://github.com/SemRels/semrel",
			Commits:       []string{"feat: add a feature"},
		}
		output := generator.Generate(ctxNoPR, opts)

		require.Contains(t, output, "made their first contribution\n")
		require.NotContains(t, output, "bob) made their first contribution in")
	})
}

func TestGeneratorMVP(t *testing.T) {
	t.Parallel()

	generator := &Generator{now: func() time.Time {
		return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	}}

	ctx := ReleaseContext{
		Version:       "1.4.0",
		RepositoryURL: "https://github.com/SemRels/semrel",
		Commits:       []string{"feat: add login (#42)", "fix: patch (#42)"},
	}

	t.Run("mvp section rendered when enabled and contributors present", func(t *testing.T) {
		t.Parallel()

		opts := DefaultGenerateOptions()
		opts.MVP = true
		opts.Contributors = []Contributor{
			{Name: "Alice", Login: "alice", PR: 42},
			{Name: "Bob", Login: "bob", PR: 99},
		}

		output := generator.Generate(ctx, opts)

		require.Contains(t, output, "### 🏆 MVP")
		require.Contains(t, output, "[@alice](https://github.com/alice)")
	})

	t.Run("mvp section not rendered when disabled", func(t *testing.T) {
		t.Parallel()

		opts := DefaultGenerateOptions()
		opts.MVP = false
		opts.Contributors = []Contributor{{Name: "Alice", Login: "alice", PR: 42}}

		output := generator.Generate(ctx, opts)

		require.NotContains(t, output, "🏆 MVP")
	})

	t.Run("mvp single contributor always wins", func(t *testing.T) {
		t.Parallel()

		mvp := pickMVP([]Contributor{{Name: "Solo", Login: "solo"}}, nil, "commits")
		require.NotNil(t, mvp)
		require.Equal(t, "solo", mvp.Login)
	})

	t.Run("mvp nil for empty contributors", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, pickMVP(nil, nil, "commits"))
	})
}

func TestFormatContributorEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		contributor   Contributor
		repositoryURL string
		want          string
	}{
		{
			name:          "login with repo url produces linked mention",
			contributor:   Contributor{Login: "alice"},
			repositoryURL: "https://github.com/SemRels/semrel",
			want:          "[@alice](https://github.com/alice)",
		},
		{
			name:          "login without repo url produces bare mention",
			contributor:   Contributor{Login: "alice"},
			repositoryURL: "",
			want:          "@alice",
		},
		{
			name:          "no login falls back to name",
			contributor:   Contributor{Name: "Alice Smith"},
			repositoryURL: "https://github.com/SemRels/semrel",
			want:          "Alice Smith",
		},
		{
			name:          "empty contributor returns unknown",
			contributor:   Contributor{},
			repositoryURL: "",
			want:          "unknown",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatContributorEntry(tt.contributor, tt.repositoryURL)
			require.Equal(t, tt.want, got)
		})
	}
}
