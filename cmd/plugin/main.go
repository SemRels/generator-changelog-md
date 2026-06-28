// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-md Authors

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	plugin "github.com/SemRels/generator-changelog-md/internal/plugin"
)

// pluginSchemaVersion is the schema version of this plugin's SEMREL_PLUGIN_* env-var contract.
const pluginSchemaVersion = 1

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Getenv, os.ReadFile, os.WriteFile))
}

func run(stdout, stderr io.Writer, getenv func(string) string, readFile func(string) ([]byte, error), writeFile func(string, []byte, os.FileMode) error) int {
	// Emit schema version so semrel core can detect config contract mismatches.
	_, _ = fmt.Fprintf(stderr, "plugin_schema_version=%d\n", pluginSchemaVersion)

	ctx, err := releaseContextFromEnv(getenv)
	if err != nil {
		fmt.Fprintln(stderr, "generator-changelog-md:", err)
		return 1
	}

	options := plugin.DefaultGenerateOptions()
	options.GroupByType = envBool(getenv, "SEMREL_PLUGIN_GROUP_BY_TYPE", true)
	options.LinkPRs = envBool(getenv, "SEMREL_PLUGIN_LINK_PRS", true)
	options.LinkCommits = envBool(getenv, "SEMREL_PLUGIN_LINK_COMMITS", true)
	options.Signature = envBool(getenv, "SEMREL_PLUGIN_SIGNATURE", false)
	options.NewContributors = envBool(getenv, "SEMREL_PLUGIN_NEW_CONTRIBUTORS", true)
	options.MVP = envBool(getenv, "SEMREL_PLUGIN_MVP", false)
	options.AIDisclosure = envBool(getenv, "SEMREL_PLUGIN_AI_DISCLOSURE", false)
	options.AIDisclosureSection = envBool(getenv, "SEMREL_PLUGIN_AI_DISCLOSURE_SECTION", false)
	if badge := strings.TrimSpace(getenv("SEMREL_PLUGIN_AI_DISCLOSURE_BADGE")); badge != "" {
		options.AIDisclosureBadge = badge
	}
	if mv := strings.TrimSpace(getenv("SEMREL_PLUGIN_MVP_METRIC")); mv != "" {
		options.MVPMetric = mv
	}

	if raw := strings.TrimSpace(getenv("SEMREL_PLUGIN_CONTRIBUTORS_JSON")); raw != "" {
		var contributors []plugin.Contributor
		if err := json.Unmarshal([]byte(raw), &contributors); err != nil {
			fmt.Fprintln(stderr, "generator-changelog-md: invalid SEMREL_PLUGIN_CONTRIBUTORS_JSON:", err)
			return 1
		}
		options.Contributors = contributors
	}

	newEntry := plugin.New().Generate(ctx, options)

	// Compression / file management.
	keepReleases := envInt(getenv, "SEMREL_PLUGIN_KEEP_RELEASES", 0)
	changelogFile := firstNonEmpty(getenv("SEMREL_PLUGIN_CHANGELOG_FILE"), "CHANGELOG.md")
	dryRun := envBool(getenv, "SEMREL_DRY_RUN", false) || envBool(getenv, "SEMREL_PLUGIN_DRY_RUN", false)

	if keepReleases > 0 {
		compressOpts := plugin.CompressOptions{
			KeepReleases: keepReleases,
			ArchiveLink:  firstNonEmpty(getenv("SEMREL_PLUGIN_ARCHIVE_LINK"), "release_url"),
			ChangelogDir: firstNonEmpty(getenv("SEMREL_PLUGIN_CHANGELOG_DIR"), "changelogs/"),
			ReleaseURL:   strings.TrimSpace(getenv("SEMREL_RELEASE_URL")),
			DryRun:       dryRun,
		}

		// Read existing changelog (ignore error if file does not exist yet).
		existingContent := ""
		if raw, readErr := readFile(changelogFile); readErr == nil {
			existingContent = string(raw)
		}

		updated, archived := plugin.CompressChangelog(existingContent, newEntry, compressOpts)

		if dryRun {
			fmt.Fprintf(stderr, "generator-changelog-md: dry-run: would write %s with %d release(s) expanded, %d archived\n",
				changelogFile, keepReleases, len(archived))
		} else {
			// Write per-release archive files when ArchiveLink includes "file".
			if compressOpts.ArchiveLink == "file" || compressOpts.ArchiveLink == "both" {
				for _, r := range archived {
					filename := plugin.ArchiveFilename(r, compressOpts.ChangelogDir)
					content := plugin.ArchiveFileContent(r)
					if err := writeFile(filename, []byte(content), 0o644); err != nil {
						fmt.Fprintf(stderr, "generator-changelog-md: warning: could not write archive file %s: %v\n", filename, err)
					}
				}
			}

			if err := writeFile(changelogFile, []byte(updated+"\n"), 0o644); err != nil {
				fmt.Fprintln(stderr, "generator-changelog-md: could not write changelog file:", err)
				return 1
			}
		}
	}

	if _, err := io.WriteString(stdout, newEntry); err != nil {
		fmt.Fprintln(stderr, "generator-changelog-md:", err)
		return 1
	}

	return 0
}

func releaseContextFromEnv(getenv func(string) string) (plugin.ReleaseContext, error) {
	raw := getenv("SEMREL_COMMITS")

	var commits []string
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &commits); err != nil {
			return plugin.ReleaseContext{}, fmt.Errorf("invalid SEMREL_COMMITS JSON: %w", err)
		}
	}

	return plugin.ReleaseContext{
		Version:        firstNonEmpty(getenv("SEMREL_VERSION"), getenv("SEMREL_TAG_NAME"), getenv("SEMREL_NEXT_VERSION")),
		CurrentVersion: strings.TrimSpace(getenv("SEMREL_CURRENT_VERSION")),
		Branch:         strings.TrimSpace(getenv("SEMREL_BRANCH")),
		RepositoryURL:  strings.TrimSpace(getenv("SEMREL_REPOSITORY_URL")),
		Commits:        commits,
	}, nil
}

func envBool(getenv func(string) string, key string, defaultValue bool) bool {
	value := strings.TrimSpace(getenv(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func envInt(getenv func(string) string, key string, defaultValue int) int {
	value := strings.TrimSpace(getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
