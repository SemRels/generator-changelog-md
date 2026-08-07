// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-md Authors

package plugin

import (
"fmt"
"regexp"
"strings"
)

// releaseHeadingPattern matches a Markdown H2 release heading such as
// "## v1.2.3 (2026-01-15)" or "## Unreleased (2026-01-15)".
var releaseHeadingPattern = regexp.MustCompile(`^## .+`)

// ArchivedRelease holds a single release entry extracted from a CHANGELOG.md.
type ArchivedRelease struct {
// Heading is the first line of the release block (e.g. "## v1.2.3 (2026-01-15)").
Heading string
// Version is extracted from the heading for display in the archive table.
Version string
// Date is extracted from the heading for display.
Date string
// Body is the raw Markdown content of the entry (without the heading).
Body string
}

// CompressOptions controls how changelog compression is applied.
type CompressOptions struct {
// KeepReleases is the number of releases to keep fully expanded. 0 means no
// compression (default — backward compatible).
KeepReleases int
// ArchiveLink controls how older entries are represented. Valid values:
//   "release_url" — append a ## Previous Releases table with links (default)
//   "file"        — write each archived entry to changelogs/vX.Y.Z.md
//   "both"        — do both
ArchiveLink string
// ChangelogDir is the directory where per-release archive files are written
// when ArchiveLink is "file" or "both". Defaults to "changelogs/".
ChangelogDir string
// ReleaseURL is the URL template for the ## Previous Releases table. Use
// SEMREL_RELEASE_URL from the provider plugin. Optional.
ReleaseURL string
// DryRun logs what would be archived without modifying any files.
DryRun bool
}

// DefaultCompressOptions returns compression options that keep the current
// (backward-compatible) behavior: no compression.
func DefaultCompressOptions() CompressOptions {
return CompressOptions{
KeepReleases: 0,
ArchiveLink:  "release_url",
ChangelogDir: "changelogs/",
}
}

// SplitReleases splits a CHANGELOG.md string into individual release blocks.
// The first block is the newest release; blocks are ordered from newest to oldest.
// Lines before the first release heading are silently ignored.
func SplitReleases(content string) []ArchivedRelease {
lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
var releases []ArchivedRelease
var current []string

flush := func() {
if len(current) == 0 {
return
}
heading := current[0]
if !releaseHeadingPattern.MatchString(heading) {
current = nil
return
}
body := strings.TrimRight(strings.Join(current[1:], "\n"), "\n")
version, date := parseHeadingVersionDate(heading)
releases = append(releases, ArchivedRelease{
Heading: heading,
Version: version,
Date:    date,
Body:    body,
})
current = nil
}

for _, line := range lines {
if releaseHeadingPattern.MatchString(line) && len(current) > 0 {
flush()
}
current = append(current, line)
}
flush()

return releases
}

// CompressChangelog applies the compression policy to existing CHANGELOG.md
// content and a newly generated entry. It returns the updated CHANGELOG.md
// content and the list of releases that were archived (moved to the table or
// separate files). When KeepReleases is 0, it simply prepends newEntry to
// existing content and returns no archived releases.
func CompressChangelog(existingContent, newEntry string, opts CompressOptions) (updated string, archived []ArchivedRelease) {
if opts.KeepReleases <= 0 {
// No compression — prepend new entry to existing file content.
return prependEntry(existingContent, newEntry), nil
}

releases := SplitReleases(strings.TrimSpace(existingContent))

// Number of releases to keep after prepending the new entry.
// We always keep the new entry plus (KeepReleases-1) existing ones.
keepExisting := opts.KeepReleases - 1
if keepExisting < 0 {
keepExisting = 0
}

var kept []ArchivedRelease
if keepExisting < len(releases) {
kept = releases[:keepExisting]
archived = releases[keepExisting:]
} else {
kept = releases
}

// Build the updated changelog: new entry + kept entries.
var sb strings.Builder
sb.WriteString(strings.TrimRight(newEntry, "\n"))
for _, r := range kept {
sb.WriteString("\n\n")
sb.WriteString(r.Heading)
if strings.TrimSpace(r.Body) != "" {
sb.WriteString("\n")
sb.WriteString(r.Body)
}
}

// Append archive table when ArchiveLink includes "release_url".
if len(archived) > 0 && (opts.ArchiveLink == "release_url" || opts.ArchiveLink == "both") {
sb.WriteString("\n\n## Previous Releases\n\n")
sb.WriteString("| Version | Date | Link |\n")
sb.WriteString("|---------|------|------|\n")
for _, r := range archived {
link := ""
if opts.ReleaseURL != "" {
link = fmt.Sprintf("[Release](%s)", opts.ReleaseURL)
}
_, _ = fmt.Fprintf(&sb, "| %s | %s | %s |\n", r.Version, r.Date, link)
}
}

return sb.String(), archived
}

// ArchiveFilename returns the filename for an archived release (used in "file" mode).
func ArchiveFilename(release ArchivedRelease, changelogDir string) string {
dir := strings.TrimRight(strings.TrimSpace(changelogDir), "/")
if dir == "" {
dir = "changelogs"
}
version := strings.TrimSpace(release.Version)
if version == "" {
version = "unknown"
}
return fmt.Sprintf("%s/%s.md", dir, version)
}

// ArchiveFileContent returns the Markdown content for an archived release file.
func ArchiveFileContent(release ArchivedRelease) string {
var sb strings.Builder
sb.WriteString(release.Heading)
if strings.TrimSpace(release.Body) != "" {
sb.WriteString("\n")
sb.WriteString(release.Body)
}
sb.WriteString("\n")
return sb.String()
}

// versionDatePattern matches "## v1.2.3 (2026-01-15)" and similar headings.
var versionDatePattern = regexp.MustCompile(`^## ([^\s(]+)\s*(?:\(([^)]+)\))?`)

func parseHeadingVersionDate(heading string) (version, date string) {
matches := versionDatePattern.FindStringSubmatch(heading)
if len(matches) >= 2 {
version = matches[1]
}
if len(matches) >= 3 {
date = matches[2]
}
return
}

func prependEntry(existing, newEntry string) string {
newEntry = strings.TrimRight(newEntry, "\n")
existing = strings.TrimLeft(existing, "\n")
if existing == "" {
return newEntry
}
return newEntry + "\n\n" + existing
}
