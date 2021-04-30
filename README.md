# go-compatible is a compatibility checker for Go packages.

[![Go Reference](https://pkg.go.dev/badge/github.com/perillo/go-compatibile.svg)](https://pkg.go.dev/github.com/perillo/go-compatibile)

## Installation

go-compatible requires [Go 1.16](https://golang.org/doc/devel/release.html#go1.16).

    go install github.com/perillo/go-compatibile@latest

## Purpose

go-compatibile checks if a package is compatible with older versions of Go.
Internally, it invokes `go vet`, `go build` or `go test` on all the
[available releases](https://pkg.go.dev/golang.org/dl) installed on the system.

The output of this tool reports problems for each release that a package does
not support.

## Usage

    go-compatible [-since goversion] [-mode mode] [packages]

Invoke `go-compatible` with one or more import paths.  go-compatible uses the
same [import path syntax](https://golang.org/cmd/go/#hdr-Import_path_syntax) as
the `go` command and therefore also supports relative import paths like
`./...`. Additionally the `...` wildcard can be used as suffix on relative and
absolute file paths to recurse into them.

The `-since` option causes the tool to only use releases more recent than the
specified version.

The `-mode` option allows the user to specify how to verify compatibility.  It
can be set to `vet`, `build` or `test`, with `vet` being the default.

By default, `go-compatible` searches the available releases in the `~/sdk`
directory, but it is possible to specify a different directory using the
`GOSDK` environment variable.
