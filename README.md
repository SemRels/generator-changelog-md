# generator-changelog-md

Markdown changelog generator plugin for Semantic Release.

Generates Markdown changelog output from Semantic Release results.

## Documentation

- Docs (coming soon): <https://github.com/SemRels/semrel/tree/main/docs/plugins/generator-changelog-md>
- Template source: <https://github.com/SemRels/plugin-template>

## Repository Layout

`	ext
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
`

## Development

`ash
go build ./cmd/plugin
go test ./...
`

## Configuration Example

`yaml
plugins:
  - name: generator-changelog-md
    type: generator
    config:
      output: CHANGELOG.md
      title: Changelog
      include_links: true
`

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.