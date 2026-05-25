// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The SemRels Authors

package plugin

import (
	"context"
	"testing"
	"time"

	semrelv1 "github.com/SemRels/semrel-api/api/gen/v1"
	"github.com/stretchr/testify/require"
)

func TestGenerateNotesGroupsFeatureAndFixCommits(t *testing.T) {
	setFixedNow(t)

	generator := New()
	response, err := generator.GenerateNotes(context.Background(), &semrelv1.GenerateNotesRequest{
		Ctx: &semrelv1.ReleaseContext{
			Commits: []*semrelv1.Commit{
				{Sha: "abc1234567", RawMessage: "feat(api): add new login endpoint"},
				{Sha: "def5678901", RawMessage: "fix: prevent crash on empty input"},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "### ✨ Features\n- feat(api): add new login endpoint (abc1234)\n\n### 🐛 Bug Fixes\n- fix: prevent crash on empty input (def5678)", response.GetNotes())
}

func TestGenerateNotesMarksBreakingChanges(t *testing.T) {
	setFixedNow(t)

	generator := New()
	response, err := generator.GenerateNotes(context.Background(), &semrelv1.GenerateNotesRequest{
		Ctx: &semrelv1.ReleaseContext{
			Commits: []*semrelv1.Commit{{
				Sha:        "abc1234567",
				RawMessage: "feat(api)!: overhaul auth flow\n\nBREAKING CHANGE: tokens issued by older releases are invalid",
			}},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "### ✨ Features\n- 💥 **BREAKING CHANGE**: feat(api): overhaul auth flow (abc1234)", response.GetNotes())
}

func TestGenerateNotesWithEmptyCommitsReturnsVersionHeader(t *testing.T) {
	setFixedNow(t)

	generator := New()
	response, err := generator.GenerateNotes(context.Background(), &semrelv1.GenerateNotesRequest{
		Ctx: &semrelv1.ReleaseContext{
			NextVersion: &semrelv1.SemanticVersion{Major: 1, Minor: 2, Patch: 3},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "## v1.2.3 (2026-05-25)", response.GetNotes())
}

func TestGenerateNotesPlacesNonConventionalCommitInOtherChanges(t *testing.T) {
	setFixedNow(t)

	generator := New()
	response, err := generator.GenerateNotes(context.Background(), &semrelv1.GenerateNotesRequest{
		Ctx: &semrelv1.ReleaseContext{
			Commits: []*semrelv1.Commit{{
				Sha:        "abc1234567",
				RawMessage: "update release docs manually\n\nextra detail",
			}},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "### 🔄 Other Changes\n- update release docs manually (abc1234)", response.GetNotes())
}

func TestGenerateNotesIncludesVersionHeaderWhenNextVersionIsSet(t *testing.T) {
	setFixedNow(t)

	generator := New()
	response, err := generator.GenerateNotes(context.Background(), &semrelv1.GenerateNotesRequest{
		Ctx: &semrelv1.ReleaseContext{
			NextVersion: &semrelv1.SemanticVersion{Major: 1, Minor: 2, Patch: 3},
			Commits: []*semrelv1.Commit{{
				Sha:        "abc1234567",
				RawMessage: "docs: refresh README examples",
			}},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "## v1.2.3 (2026-05-25)\n\n### 📚 Documentation\n- docs: refresh README examples (abc1234)", response.GetNotes())
}

func setFixedNow(t *testing.T) {
	t.Helper()

	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	}

	t.Cleanup(func() {
		nowFunc = originalNowFunc
	})
}
