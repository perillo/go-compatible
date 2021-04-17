// Copyright 2021 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/perillo/go-compatible/internal/invoke"
	"github.com/perillo/go-compatible/internal/version"
)

// gosdk is the path to go sdk directory, by default ~/sdk.  It can be
// overridden using the GOSDK environment variable.
var gosdk string

type release struct {
	goroot  string
	version version.Version
}

func (r release) String() string {
	return "go" + r.version.String()
}

func init() {
	if value, ok := os.LookupEnv("GOSDK"); ok {
		gosdk = value

		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get home directory: %v\n", err)

		return
	}
	gosdk = filepath.Join(home, "sdk")
}

func main() {
	// Setup log.
	log.SetFlags(0)

	// Parse command line.
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Usage: go-compatible [flags] packages")
		fmt.Fprintf(w, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()

	releases, err := gosdklist()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(releases, args); err != nil {
		log.Fatal(err)
	}
}

// run invokes go vet for all the specified releases.
func run(releases []release, patterns []string) error {
	nl := []byte("\n")
	index := 0 // current failed release

	for _, rel := range releases {
		msg, err := govet(rel, patterns)
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}

		// Print go vet diagnostic message.
		if index > 0 {
			os.Stderr.Write(nl)
		}
		fmt.Fprintf(os.Stderr, "using go%s\n", rel.version)
		os.Stderr.Write(msg)
		os.Stderr.Write(nl)

		index++
	}

	return nil
}

// gosdklist returns a list of all go releases in the sdk.
func gosdklist() ([]release, error) {
	list := make([]release, 0, 32) // preallocate memory
	files, err := os.ReadDir(gosdk)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		name := file.Name()
		if file.IsDir() && strings.HasPrefix(name, "go") {
			goroot := filepath.Join(gosdk, name)
			line, err := goversion(goroot)
			if err != nil {
				return nil, err
			}
			version, err := version.ParseLine(line)
			if err != nil {
				return nil, err
			}

			rel := release{
				goroot:  goroot,
				version: version,
			}
			list = append(list, rel)
		}
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no go releases found in %s", gosdk)
	}

	// Sort the releases.
	sort.Slice(list, func(i, j int) bool {
		return list[i].version.Less(list[j].version)
	})

	return list, nil
}

// goversion returns the version of go from goroot.
func goversion(goroot string) (string, error) {
	gocmd := filepath.Join(goroot, "bin", "go")
	cmd := exec.Command(gocmd, "version")
	cmd.Env = append(os.Environ(), "GOROOT="+goroot)

	stdout, err := invoke.Output(cmd)
	if err != nil {
		// TODO(mperillo): Ignore the case of gocmd not found.
		return "", err
	}

	return string(stdout), nil
}

// govet invokes go vet on the packages named by the given patterns, for the
// specified release.  It returns the diagnostic message and a non nil error,
// in case of a fatal error like go command not found or incorrect command line
// arguments.
func govet(rel release, patterns []string) ([]byte, error) {
	// TODO(mperillo): go1.4 does not have the go vet tool;  report an useful
	// error if the user has not installed it.
	gocmd := filepath.Join(rel.goroot, "bin", "go")
	args := append([]string{"vet"}, patterns...)
	cmd := exec.Command(gocmd, args...)
	cmd.Env = append(os.Environ(), "GOROOT="+rel.goroot)

	if err := invoke.Run(cmd); err != nil {
		cmderr := err.(*invoke.Error)

		// Determine the error type to decide if there was a fatal problem
		// with the invocation of go vet that requires the termination of
		// the program.
		switch cmderr.Err.(type) {
		case *exec.Error:
			return nil, err
		case *exec.ExitError:
			if isFatal(cmderr) {
				return nil, err
			}

			return cmderr.Stderr, nil
		}
	}

	return nil, nil
}

// isFatal returns true if the error returned by go vet is fatal.
func isFatal(err *invoke.Error) bool {
	// In case of build constraints excluding all Go files, go vet returns
	// exit status 1 and the error message starts with "package".
	//
	// TODO(mperillo): all Go files excluded due to build constraints is
	// probably a fatal error.
	if bytes.HasPrefix(err.Stderr, []byte("package")) {
		return false
	}

	// In case of syntax errors, go vet returns exit status 2 and the error
	// message starts with # and the package name.
	if err.Stderr[0] == '#' {
		return false
	}

	return false
}
