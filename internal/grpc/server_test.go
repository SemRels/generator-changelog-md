// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-semrel Authors

package grpc

import (
	"context"
	"testing"

	semrelv1 "github.com/SemRels/generator-changelog-md/internal/gen/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNotes_WithCommits(t *testing.T) {
	t.Parallel()

	srv := NewChangelogServer()
	req := &semrelv1.GenerateNotesRequest{
		Ctx: &semrelv1.ReleaseContext{
			NextVersion: &semrelv1.SemanticVersion{Major: 1, Minor: 2, Patch: 0},
			Commits: []*semrelv1.Commit{
				{Sha: "aaa", RawMessage: "feat: new API"},
				{Sha: "bbb", RawMessage: "fix: memory leak"},
			},
		},
	}

	resp, err := srv.GenerateNotes(context.Background(), req)
	require.NoError(t, err)
	assert.Contains(t, resp.GetNotes(), "## [v1.2.0]")
	assert.Contains(t, resp.GetNotes(), "new API")
	assert.Contains(t, resp.GetNotes(), "memory leak")
}

func TestGenerateNotes_NilContext(t *testing.T) {
	t.Parallel()

	srv := NewChangelogServer()
	resp, err := srv.GenerateNotes(context.Background(), &semrelv1.GenerateNotesRequest{})
	require.NoError(t, err)
	assert.Equal(t, "", resp.GetNotes())
}
