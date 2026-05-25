# generator-changelog-md

Markdown changelog generator plugin for [SemRel](https://github.com/SemRels/semrel).

Implements the `ChangelogGeneratorPlugin` gRPC service defined in the
[SemRel protobuf contract](proto/v1/semantic_release.proto) (ADR-001).

## Features

- Parses [Conventional Commits](https://www.conventionalcommits.org/) from release context
- Groups commits into **Breaking Changes**, **Features**, **Fixes**, and **Other Changes**
- Formats output as `## [vX.Y.Z] - YYYY-MM-DD` Markdown
- Helper utilities for prepending or appending new sections to an existing `CHANGELOG.md`
- Pure gRPC server — no shared-memory plugin ABI required

## Repository Layout

~~~text
cmd/plugin/              gRPC server entry point
internal/plugin/         Business logic (parsing, grouping, formatting)
internal/grpc/           gRPC transport adapter
internal/gen/v1/         Generated protobuf/gRPC stubs (do not edit)
proto/v1/                Vendored protobuf contract
~~~

## Configuration

In `.semrel.yaml` pass config keys through the `config` map of `ReleaseContext`:

| Key | Default | Description |
|-----|---------|-------------|
| *(none currently)* | | Future: `output_file`, `prepend`, `section_order` |

Example plugin block:

~~~yaml
plugins:
  - name: generator-changelog-md
    type: generator
    config:
      output_file: CHANGELOG.md
      prepend: true
      section_order:
        - Breaking Changes
        - Features
        - Fixes
~~~

## Running

~~~bash
# Start the gRPC server on the default port :50051
go run ./cmd/plugin

# Override the listen address
SEMREL_PLUGIN_ADDR=":9090" go run ./cmd/plugin
~~~

Output (stderr only):

~~~
[generator-changelog-md] 2026/05/25 10:40:00 gRPC server listening on [::]:50051
~~~

## Development

~~~bash
# Build
make build           # → bin/plugin

# Test with coverage
make test            # or: go test -v -cover ./...

# Lint (requires golangci-lint)
make lint
~~~

## Example Output

Given commits:

~~~
feat(auth): add OAuth2 login
fix: correct pagination off-by-one
feat!: remove v1 API
chore: bump Go version
~~~

`GenerateNotes` returns:

~~~markdown
## [v2.0.0] - 2026-05-25

### ⚠ Breaking Changes

- remove v1 API

### Features

- **auth**: add OAuth2 login

### Fixes

- correct pagination off-by-one

### Other Changes

- bump Go version
~~~

## Error Scenarios

| Scenario | Behaviour |
|----------|-----------|
| `nil` `ReleaseContext` | Returns empty string; no error |
| Non-conventional commit message | Treated as "Other Changes" |
| `BREAKING CHANGE:` footer in body | Commit moved to Breaking Changes section |
| `feat!:` bang syntax | Commit moved to Breaking Changes section |
| No commits | Header line only; no section headings |

## Links

- [SemRel ADR-001 – gRPC plugin transport](https://github.com/SemRels/semrel/blob/main/docs/adr/ADR-001-grpc-plugin-transport.md)
- [Plugin template](https://github.com/SemRels/plugin-template)
- [Registry](https://registry.semrel.io)

