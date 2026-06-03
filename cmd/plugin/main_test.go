// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-md Authors

package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunWritesMarkdown(t *testing.T) {
	t.Parallel()

	getenv := func(key string) string {
		switch key {
		case "SEMREL_VERSION":
			return "1.3.0"
		case "SEMREL_REPOSITORY_URL":
			return "https://github.com/SemRels/semrel"
		case "SEMREL_COMMITS":
			return `["feat: add search (#123)"]`
		}
		return ""
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(&stdout, &stderr, getenv)

	require.Equal(t, 0, code)
	require.Empty(t, stderr.String())
	require.Contains(t, stdout.String(), "## v1.3.0")
	require.Contains(t, stdout.String(), "### Features")
	require.Contains(t, stdout.String(), "[#123](https://github.com/SemRels/semrel/pull/123)")
}

func TestRunWritesFlatMarkdownWhenGroupingDisabled(t *testing.T) {
	t.Parallel()

	getenv := func(key string) string {
		switch key {
		case "SEMREL_VERSION":
			return "1.3.0"
		case "SEMREL_PLUGIN_GROUP_BY_TYPE":
			return "false"
		case "SEMREL_COMMITS":
			return `["feat: add search","fix: resolve crash"]`
		}
		return ""
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(&stdout, &stderr, getenv)

	require.Equal(t, 0, code)
	require.Empty(t, stderr.String())
	require.NotContains(t, stdout.String(), "### Features")
	require.Contains(t, stdout.String(), "\n- feat: add search\n- fix: resolve crash")
}

func TestRunRejectsInvalidCommitJSON(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(&stdout, &stderr, func(key string) string {
		if key == "SEMREL_COMMITS" {
			return `[`
		}
		return ""
	})

	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "invalid SEMREL_COMMITS JSON")
}

func TestReleaseContextFromEnvUsesVersionFallback(t *testing.T) {
	t.Parallel()

	ctx, err := releaseContextFromEnv(func(key string) string {
		switch key {
		case "SEMREL_TAG_NAME":
			return "v2.0.0"
		case "SEMREL_CURRENT_VERSION":
			return "1.9.0"
		case "SEMREL_BRANCH":
			return "release/main"
		case "SEMREL_REPOSITORY_URL":
			return "https://github.com/SemRels/semrel"
		}
		return ""
	})

	require.NoError(t, err)
	require.Equal(t, "v2.0.0", ctx.Version)
	require.Equal(t, "1.9.0", ctx.CurrentVersion)
	require.Equal(t, "release/main", ctx.Branch)
	require.Equal(t, "https://github.com/SemRels/semrel", ctx.RepositoryURL)
}
