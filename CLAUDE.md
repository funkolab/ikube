# ikube Development Guide

## Build/Test Commands
- Build: `task build`
- Install: `task install`
- Run tests: `task test`
- Run specific test: `go test -v ./... -run TestFunctionName`
- Lint: `task lint` (uses golangci-lint)
- Vet: `task vet`
- Static check: `task check`
- Run all (lint, test, build): `task all`

## Code Style Guidelines
- **Error handling**: Wrap errors with context using `fmt.Errorf("message: %v", err)`
- **Naming**: Use camelCase for variables/functions, UPPER_CASE for constants
- **Functions**: Prefix handler functions with `handle` (e.g., `handleListSecrets`)
- **Imports**: Standard library first, then third-party packages
- **Formatting**: Use tabs for indentation, empty lines between logical sections
- **Documentation**: Add comments for functions explaining their purpose
- **Error output**: Show detailed errors in verbose mode, simpler messages otherwise
- **Input validation**: Validate early with specific error messages
- **Resource cleanup**: Use `defer` for proper cleanup of resources

For detailed Go guidelines, follow [Effective Go](https://golang.org/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).