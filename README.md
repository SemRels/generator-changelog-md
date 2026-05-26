# generator-changelog-md

Generates a Markdown changelog for the current release.

This plugin is distributed as the standalone Go binary `semrel-plugin-generator-changelog-md`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

```bash
go install github.com/SemRels/generator-changelog-md/cmd/plugin@latest
```

## Configuration

```yaml
plugins:
  - name: generator-changelog-md
    path: ~/.semrel/plugins/semrel-plugin-generator-changelog-md
    env:
      SEMREL_PLUGIN_TEMPLATE: ".semrel/templates/changelog.md.tmpl"
      SEMREL_PLUGIN_MAX_COMMITS: "100"
```

## `SEMREL_PLUGIN_*` variables

| Name | Required | Description | Default |
| --- | --- | --- | --- |
| `SEMREL_PLUGIN_TEMPLATE` | Optional | Path to a custom Go template file. | Built-in template |
| `SEMREL_PLUGIN_MAX_COMMITS` | Optional | Maximum number of commits to include. | 100 |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CURRENT_VERSION` | Current version before the new release is applied. |
| `SEMREL_BRANCH` | Git branch associated with the current release run. |

## Example behavior

The plugin renders a Markdown changelog from the release context and commit history, then prints the changelog so semrel can pass it to later plugins.

## License

Apache-2.0
