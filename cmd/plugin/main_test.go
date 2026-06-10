// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-md Authors

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func noopFS() (func(string) ([]byte, error), func(string, []byte, os.FileMode) error) {
	readFile := func(string) ([]byte, error) { return nil, os.ErrNotExist }
	writeFile := func(string, []byte, os.FileMode) error { return nil }
	return readFile, writeFile
}

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
	rf, wf := noopFS()

	code := run(&stdout, &stderr, getenv, rf, wf)

	require.Equal(t, 0, code)
	require.Contains(t, stderr.String(), "plugin_schema_version=1")
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
	rf, wf := noopFS()

	code := run(&stdout, &stderr, getenv, rf, wf)

	require.Equal(t, 0, code)
	require.Contains(t, stderr.String(), "plugin_schema_version=1")
	require.NotContains(t, stdout.String(), "### Features")
	require.Contains(t, stdout.String(), "\n- feat: add search\n- fix: resolve crash")
}

func TestRunRejectsInvalidCommitJSON(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	rf, wf := noopFS()

	code := run(&stdout, &stderr, func(key string) string {
		if key == "SEMREL_COMMITS" {
			return `[`
		}
		return ""
	}, rf, wf)

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

func TestRunWithCompressionWritesChangelog(t *testing.T) {
	t.Parallel()

	existing := `## v1.2.0 (2026-04-01)

### Features
- feat: old feature

## v1.1.0 (2026-03-01)

### Bug Fixes
- fix: old fix
`
	getenv := func(key string) string {
		switch key {
		case "SEMREL_VERSION":
			return "1.3.0"
		case "SEMREL_COMMITS":
			return `["feat: new feature"]`
		case "SEMREL_PLUGIN_KEEP_RELEASES":
			return "2"
		case "SEMREL_PLUGIN_CHANGELOG_FILE":
			return "CHANGELOG.md"
		}
		return ""
	}

	var writtenPath string
	var writtenContent []byte

	readFile := func(string) ([]byte, error) { return []byte(existing), nil }
	writeFile := func(path string, data []byte, _ os.FileMode) error {
		writtenPath = path
		writtenContent = data
		return nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(&stdout, &stderr, getenv, readFile, writeFile)

	require.Equal(t, 0, code)
	require.Contains(t, stderr.String(), "plugin_schema_version=1")

	// stdout contains only the new entry (for release notes).
	require.Contains(t, stdout.String(), "## v1.3.0")
	require.NotContains(t, stdout.String(), "v1.1.0")

	// The CHANGELOG.md file contains new + 1 kept + archived table.
	require.Equal(t, "CHANGELOG.md", writtenPath)
	content := string(writtenContent)
	require.Contains(t, content, "## v1.3.0")
	require.Contains(t, content, "## v1.2.0")
	require.Contains(t, content, "## Previous Releases")
	require.Contains(t, content, "v1.1.0")
	require.NotContains(t, content, "### Bug Fixes\n- fix: old fix")
}

func TestRunWithCompressionDryRunSkipsWrite(t *testing.T) {
	t.Parallel()

	getenv := func(key string) string {
		switch key {
		case "SEMREL_VERSION":
			return "1.3.0"
		case "SEMREL_COMMITS":
			return `["feat: new feature"]`
		case "SEMREL_PLUGIN_KEEP_RELEASES":
			return "1"
		case "SEMREL_DRY_RUN":
			return "true"
		}
		return ""
	}

	written := false
	readFile := func(string) ([]byte, error) { return []byte("## v1.2.0 (2026-04-01)\n\n- old"), nil }
	writeFile := func(string, []byte, os.FileMode) error { written = true; return nil }

	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, getenv, readFile, writeFile)

	require.Equal(t, 0, code)
	require.False(t, written)
	require.Contains(t, stderr.String(), "dry-run")
}
