# highway-to-golang

Learning project organized as a small Go task application.

## Layout

```text
cmd/task                 application entrypoint
internal/app             application wiring and demo scenario
internal/task            task domain, service, errors, storage interface
internal/task/storage    storage implementations
```

The layout is inspired by common Go project structure patterns, but keeps only
the directories that are useful for this project.

## Commands

```bash
go generate ./...
go test ./...
go test -race ./...
go run ./cmd/task
```
