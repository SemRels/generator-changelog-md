// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-semrel Authors

package plugin

import (
	"strings"
	"testing"

	semrelv1 "github.com/SemRels/generator-changelog-md/internal/gen/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ParseCommit
// ---------------------------------------------------------------------------

func TestParseCommit_ConventionalFeat(t *testing.T) {
	t.Parallel()
	pc := ParseCommit("abc123", "feat(auth): add OAuth2 login")
	assert.Equal(t, "feat", pc.Type)
	assert.Equal(t, "auth", pc.Scope)
	assert.Equal(t, "add OAuth2 login", pc.Description)
	assert.False(t, pc.Breaking)
	assert.Equal(t, "abc123", pc.SHA)
}

func TestParseCommit_ConventionalFix(t *testing.T) {
	t.Parallel()
	pc := ParseCommit("def456", "fix: correct off-by-one in pagination")
	assert.Equal(t, "fix", pc.Type)
	assert.Equal(t, "", pc.Scope)
	assert.Equal(t, "correct off-by-one in pagination", pc.Description)
}

func TestParseCommit_BreakingBang(t *testing.T) {
	t.Parallel()
	pc := ParseCommit("ghi789", "feat!: remove legacy API")
	assert.True(t, pc.Breaking)
	assert.Equal(t, "feat", pc.Type)
}

func TestParseCommit_BreakingFooter(t *testing.T) {
	t.Parallel()
	msg := "refactor: drop support for v1\n\nBREAKING CHANGE: v1 endpoints removed"
	pc := ParseCommit("jkl000", msg)
	assert.True(t, pc.Breaking)
}

func TestParseCommit_NonConventional(t *testing.T) {
	t.Parallel()
	pc := ParseCommit("zzz999", "Update README")
	assert.Equal(t, "", pc.Type)
	assert.Equal(t, "Update README", pc.Description)
	assert.False(t, pc.Breaking)
}

// ---------------------------------------------------------------------------
// GroupCommits
// ---------------------------------------------------------------------------

func makeCommits(msgs ...string) []*semrelv1.Commit {
	out := make([]*semrelv1.Commit, len(msgs))
	for i, m := range msgs {
		out[i] = &semrelv1.Commit{Sha: "sha" + string(rune('0'+i)), RawMessage: m}
	}
	return out
}

func TestGroupCommits_Basic(t *testing.T) {
	t.Parallel()
	commits := makeCommits(
		"feat: new dashboard",
		"fix: broken import",
		"chore: update deps",
		"feat!: drop v1",
	)
	g := GroupCommits(commits)
	require.Len(t, g.Features, 1)
	require.Len(t, g.Fixes, 1)
	require.Len(t, g.Others, 1)
	require.Len(t, g.BreakingChanges, 1)
	assert.Equal(t, "new dashboard", g.Features[0].Description)
}

func TestGroupCommits_Empty(t *testing.T) {
	t.Parallel()
	g := GroupCommits(nil)
	assert.Empty(t, g.Features)
	assert.Empty(t, g.Fixes)
	assert.Empty(t, g.BreakingChanges)
	assert.Empty(t, g.Others)
}

func TestGroupCommits_NilEntry(t *testing.T) {
	t.Parallel()
	commits := []*semrelv1.Commit{nil, {Sha: "a", RawMessage: "fix: nil guard"}}
	g := GroupCommits(commits)
	require.Len(t, g.Fixes, 1)
}

func TestGroupCommits_AllBreaking(t *testing.T) {
	t.Parallel()
	commits := makeCommits(
		"feat!: breaking one",
		"fix!: breaking two",
		"chore: not breaking\n\nBREAKING CHANGE: surprise",
	)
	g := GroupCommits(commits)
	assert.Len(t, g.BreakingChanges, 3)
	assert.Empty(t, g.Features)
	assert.Empty(t, g.Fixes)
}

// ---------------------------------------------------------------------------
// FormatMarkdown
// ---------------------------------------------------------------------------

func makeCtx(major, minor, patch uint32, msgs ...string) *semrelv1.ReleaseContext {
	return &semrelv1.ReleaseContext{
		NextVersion: &semrelv1.SemanticVersion{Major: major, Minor: minor, Patch: patch},
		Commits:     makeCommits(msgs...),
	}
}

func TestFormatMarkdown_Header(t *testing.T) {
	t.Parallel()
	md := FormatMarkdown(makeCtx(1, 2, 3))
	assert.True(t, strings.HasPrefix(md, "## [v1.2.3]"), "header: %q", md)
}

func TestFormatMarkdown_NilContext(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", FormatMarkdown(nil))
}

func TestFormatMarkdown_Features(t *testing.T) {
	t.Parallel()
	md := FormatMarkdown(makeCtx(0, 1, 0, "feat(ui): dark mode"))
	assert.Contains(t, md, "### Features")
	assert.Contains(t, md, "**ui**: dark mode")
}

func TestFormatMarkdown_Fixes(t *testing.T) {
	t.Parallel()
	md := FormatMarkdown(makeCtx(0, 0, 1, "fix: correct typo"))
	assert.Contains(t, md, "### Fixes")
	assert.Contains(t, md, "correct typo")
}

func TestFormatMarkdown_BreakingChanges(t *testing.T) {
	t.Parallel()
	md := FormatMarkdown(makeCtx(2, 0, 0, "feat!: remove v1 API", "feat: normal"))
	assert.Contains(t, md, "### ⚠ Breaking Changes")
	assert.Contains(t, md, "remove v1 API")
	// Breaking commits are removed from Features section.
	assert.NotContains(t, md, "remove v1 API\n- normal")
}

func TestFormatMarkdown_EmptyCommits(t *testing.T) {
	t.Parallel()
	md := FormatMarkdown(makeCtx(1, 0, 0))
	assert.True(t, strings.HasPrefix(md, "## [v1.0.0]"))
	assert.NotContains(t, md, "###")
}

func TestFormatMarkdown_Others(t *testing.T) {
	t.Parallel()
	md := FormatMarkdown(makeCtx(0, 0, 1, "chore: bump go"))
	assert.Contains(t, md, "### Other Changes")
}

// ---------------------------------------------------------------------------
// PrependToChangelog / AppendToChangelog
// ---------------------------------------------------------------------------

func TestPrependToChangelog_EmptyExisting(t *testing.T) {
	t.Parallel()
	result := PrependToChangelog("## [v1.0.0] - 2026-01-01\n\n- feat: init", "")
	assert.True(t, strings.HasPrefix(result, "# Changelog"))
	assert.Contains(t, result, "## [v1.0.0]")
}

func TestPrependToChangelog_WithTitle(t *testing.T) {
	t.Parallel()
	existing := "# Changelog\n\n## [v0.9.0] - 2025-12-01\n\n- fix: old"
	result := PrependToChangelog("## [v1.0.0] - 2026-01-01\n\n- feat: new", existing)
	// New section must appear before old section.
	idxNew := strings.Index(result, "## [v1.0.0]")
	idxOld := strings.Index(result, "## [v0.9.0]")
	require.Greater(t, idxOld, idxNew, "new section should come before old")
}

func TestPrependToChangelog_NoTitle(t *testing.T) {
	t.Parallel()
	existing := "## [v0.1.0] - 2024-01-01\n\n- chore: initial"
	result := PrependToChangelog("## [v1.0.0] - 2026-01-01\n\n- feat: new", existing)
	assert.True(t, strings.HasPrefix(result, "## [v1.0.0]"))
}

func TestAppendToChangelog_EmptyExisting(t *testing.T) {
	t.Parallel()
	result := AppendToChangelog("## [v1.0.0] - 2026-01-01\n\n- feat: init", "")
	assert.Contains(t, result, "# Changelog")
	assert.Contains(t, result, "## [v1.0.0]")
}

func TestAppendToChangelog_WithExisting(t *testing.T) {
	t.Parallel()
	existing := "# Changelog\n\n## [v0.9.0] - 2025-12-01\n\n- fix: old"
	result := AppendToChangelog("## [v1.0.0] - 2026-01-01\n\n- feat: new", existing)
	idxNew := strings.Index(result, "## [v1.0.0]")
	idxOld := strings.Index(result, "## [v0.9.0]")
	require.Greater(t, idxNew, idxOld, "new section should come after old")
}
