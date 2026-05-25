# generator-changelog-md

`generator-changelog-md` is a SemRels changelog generator plugin implemented as a go-plugin gRPC binary.

It reads commits from the SemRels `ReleaseContext`, parses Conventional Commit headers, groups entries into Markdown sections, and renders release notes such as:

```markdown
## v1.2.3 (2026-05-25)

### ✨ Features
- feat(api): add new login endpoint (abc1234)

### 🐛 Bug Fixes
- fix: prevent crash on empty input (def5678)
```

## Behavior

- Supports Conventional Commits using the SemRels analyzer regex.
- Groups `feat`, `fix`/`revert`, `perf`, and `docs` into dedicated sections.
- Sends other commit types and non-conventional commits to `🔄 Other Changes`.
- Marks breaking changes with `💥 **BREAKING CHANGE**:`.
- Uses the next semantic version from `ReleaseContext` when available.

## Development

```bash
go build ./...
go test ./...
```

## Repository Layout

```text
cmd/plugin/              go-plugin entry point
internal/plugin/         changelog generation logic and tests
internal/grpc/           reserved for future transport-related code
proto/v1                 SemRels protobuf contract
```
