# generator-changelog-md

Markdown changelog generator plugin for SemRel.

Generates release notes and changelog entries in Markdown format.

## Documentation

- SemRel docs (planned): <https://github.com/SemRels/semrel/tree/main/docs/plugins/generator-changelog-md>
- Plugin template: <https://github.com/SemRels/plugin-template>
- Registry: <https://registry.semrel.io>

## Repository Layout

~~~text
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
~~~

## Development

~~~bash
go build ./cmd/plugin
go test ./...
~~~

## Configuration Example

~~~yaml
plugins:
  - name: generator-changelog-md
    type: generator
    config:
      output_file: CHANGELOG.md
      include_compare_link: true
      section_order:
        - Features
        - Fixes
        - Documentation
~~~

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.
