# go

[![Go Reference](https://pkg.go.dev/badge/github.com/mrlm-net/go.svg)](https://pkg.go.dev/github.com/mrlm-net/go)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrlm-net/go)](https://goreportcard.com/report/github.com/mrlm-net/go)

A collection of small, focused Go packages solving common problems not covered by the standard library.

## Requirements

- Go 1.25+

## Installation

```bash
go get github.com/mrlm-net/go
```

## Packages

| Package | Description |
|---------|-------------|
| *Coming soon* | |

## Philosophy

This project follows [Go Proverbs](https://go-proverbs.github.io/):

- Clear is better than clever
- Make the zero value useful
- The bigger the interface, the weaker the abstraction
- A little copying is better than a little dependency
- Errors are values

## Development

```bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

## License

See [LICENSE](LICENSE) for details.
