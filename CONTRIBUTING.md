# Contributing Guidelines

Pull requests and issues are welcome. Contributions must adhere to the architectural invariants and coding standards below.

## Code Standards

| Invariant | Requirement |
| :--- | :--- |
| **Dependencies** | Zero external dependencies. All implementations must use the Go standard library and gda exclusively. |
| **Go Version** | Go 1.24 or higher. |
| **Comments** | Zero comments in production Go files. Code must be self-documenting through clear naming. |
| **Formatting** | All Go files must be formatted with standard `gofmt`. |
| **Concurrency Safety** | All tests must pass cleanly under the Go race detector. |

## Verification Checklist

Before opening a pull request, run local verification:

```bash
gofmt -l .
go vet ./...
go test -race -count=1 ./...
```

All automated checks must pass before a pull request can be merged.

## Commit Guidelines

Use concise, direct commit messages without prefixes, capitalization flags or emojis
- `add feature description`
- `fix defect description`
- `update component description`

