// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-md Authors

package plugin

import (
"testing"

"github.com/stretchr/testify/require"
)

func TestSplitReleases(t *testing.T) {
t.Parallel()

content := `## v1.3.0 (2026-05-01)

### Features
- feat: new feature

## v1.2.0 (2026-04-01)

### Bug Fixes
- fix: old fix

## v1.1.0 (2026-03-01)

- chore: initial`

releases := SplitReleases(content)
require.Len(t, releases, 3)
require.Equal(t, "## v1.3.0 (2026-05-01)", releases[0].Heading)
require.Equal(t, "v1.3.0", releases[0].Version)
require.Equal(t, "2026-05-01", releases[0].Date)
require.Contains(t, releases[0].Body, "feat: new feature")
require.Equal(t, "## v1.2.0 (2026-04-01)", releases[1].Heading)
require.Equal(t, "## v1.1.0 (2026-03-01)", releases[2].Heading)
}

func TestSplitReleasesEmpty(t *testing.T) {
t.Parallel()

require.Empty(t, SplitReleases(""))
require.Empty(t, SplitReleases("   "))
}

func TestCompressChangelogNoCompression(t *testing.T) {
t.Parallel()

opts := CompressOptions{KeepReleases: 0}
existing := "## v1.2.0 (2026-04-01)\n\n- old entry"
newEntry := "## v1.3.0 (2026-05-01)\n\n- new entry"

updated, archived := CompressChangelog(existing, newEntry, opts)

require.Empty(t, archived)
require.Contains(t, updated, "## v1.3.0")
require.Contains(t, updated, "## v1.2.0")
}

func TestCompressChangelogKeep1(t *testing.T) {
t.Parallel()

existing := `## v1.2.0 (2026-04-01)

### Features
- feat: old feature

## v1.1.0 (2026-03-01)

### Bug Fixes
- fix: very old fix`

newEntry := "## v1.3.0 (2026-05-01)\n\n### Features\n- feat: brand new"

opts := CompressOptions{
KeepReleases: 1,
ArchiveLink:  "release_url",
ReleaseURL:   "https://github.com/example/repo/releases/tag/v1.2.0",
}

updated, archived := CompressChangelog(existing, newEntry, opts)

require.Len(t, archived, 2)
require.Equal(t, "v1.2.0", archived[0].Version)
require.Equal(t, "v1.1.0", archived[1].Version)
require.Contains(t, updated, "## v1.3.0")
require.NotContains(t, updated, "## v1.2.0 (2026-04-01)\n")
require.Contains(t, updated, "## Previous Releases")
require.Contains(t, updated, "v1.2.0")
require.Contains(t, updated, "v1.1.0")
}

func TestCompressChangelogKeep2(t *testing.T) {
t.Parallel()

existing := `## v1.2.0 (2026-04-01)

- old

## v1.1.0 (2026-03-01)

- older

## v1.0.0 (2026-02-01)

- oldest`

newEntry := "## v1.3.0 (2026-05-01)\n\n- new"

opts := CompressOptions{
KeepReleases: 2,
ArchiveLink:  "release_url",
}

updated, archived := CompressChangelog(existing, newEntry, opts)

require.Len(t, archived, 2, "v1.1.0 and v1.0.0 should be archived")
require.Contains(t, updated, "## v1.3.0")
require.Contains(t, updated, "## v1.2.0")
require.Contains(t, updated, "## Previous Releases")
require.Contains(t, updated, "v1.1.0")
require.Contains(t, updated, "v1.0.0")
require.NotContains(t, updated, "- older")
}

func TestCompressChangelogNoExistingContent(t *testing.T) {
t.Parallel()

opts := CompressOptions{KeepReleases: 3}
newEntry := "## v1.0.0 (2026-01-01)\n\n- initial"

updated, archived := CompressChangelog("", newEntry, opts)

require.Empty(t, archived)
require.Contains(t, updated, "## v1.0.0")
}

func TestArchiveFilename(t *testing.T) {
t.Parallel()

r := ArchivedRelease{Version: "v1.2.3"}
require.Equal(t, "changelogs/v1.2.3.md", ArchiveFilename(r, "changelogs/"))
require.Equal(t, "archive/v1.2.3.md", ArchiveFilename(r, "archive"))
}

func TestArchiveFileContent(t *testing.T) {
t.Parallel()

r := ArchivedRelease{
Heading: "## v1.2.3 (2026-01-01)",
Body:    "\n### Features\n- feat: something",
}
content := ArchiveFileContent(r)
require.Contains(t, content, "## v1.2.3")
require.Contains(t, content, "feat: something")
}
