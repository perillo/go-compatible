// Copyright 2021 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package version provides support for parsing and sorting go versions.
package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// regex is based on the semver regex from https://regex101.com/r/Ly7O1x/3/.
var regex = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)(?:\.(?P<patch>0|[1-9]\d*))?(?P<prerelease>.*)$`)

// Version represents a Go version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
}

// ParseLine parses the version line returned by go version.
func ParseLine(line string) (Version, error) {
	// The line returned by go version for stable releases is:
	//   "go version go<version> <os>/<arch>"
	// For unstable releases it is:
	//   "go version devel go<version> <timestamp> <os>/<arch>"
	fields := strings.Fields(line)
	version := fields[2] // field after "go version"
	if version == "devel" {
		version = fields[3] // field after "go version devel"
	}

	return Parse(version)
}

// Parse parses the Go version.
func Parse(version string) (v Version, err error) {
	// ABNF for Go version:
	//   "go" major "." minor ["." patch] [pre-release]
	//
	// As an example:
	// go1.16
	// go1.16.3
	// go1.16beta1
	// go1.17-3f4977bd58
	//
	// Note that "-" is considered part of the pre-release.
	if !strings.HasPrefix(version, "go") {
		return v, fmt.Errorf("parse: version does not have the \"go\" prefix")
	}
	version = version[2:] // strip the "go" prefix

	r := regex.FindAllStringSubmatch(version, -1)
	if r == nil {
		return v, fmt.Errorf("parse: unable to parse version %s", version)
	}
	m := r[0]

	major, err := strconv.Atoi(m[1])
	if err != nil {
		return v, fmt.Errorf("parse: invalid major in version %s", version)
	}
	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return v, fmt.Errorf("parse: invalid minor in version %s", version)
	}
	patch := 0
	if m[3] != "" {
		patch, err = strconv.Atoi(m[3])
		if err != nil {
			return v, fmt.Errorf("parse: invalid patch in version %s", version)
		}
	}
	v = Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: m[4],
	}

	return v, nil
}

// Compare returns an integer comparing two versions according to version
// precedence.
// The result will be 0 if v == w, -1 if v < w, or +1 if v > w.
func (v Version) Compare(w Version) int {
	if c := intcmp(v.Major, w.Major); c != 0 {
		return c
	}
	if c := intcmp(v.Minor, w.Minor); c != 0 {
		return c
	}
	if c := intcmp(v.Patch, w.Patch); c != 0 {
		return c
	}

	return precmp(v.PreRelease, w.PreRelease)
}

// Less returns true if v < w according to version precedence.
func (v Version) Less(w Version) bool {
	return v.Compare(w) < 0
}

// String implements the Stringer interface.
func (v Version) String() string {
	s := strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor)
	if v.Patch > 0 {
		s += "." + strconv.Itoa(v.Patch)
	}
	if v.PreRelease != "" {
		s += v.PreRelease
	}

	return s
}

// intcmp compares two integers.
func intcmp(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}

	return 0
}

// strcmp compares two strings.
func strcmp(x, y string) int {
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	}

	return 0
}

// precmp compare two pre-releases.
func precmp(x, y string) int {
	switch {
	case x == y:
		return 0
	case x == "":
		return 1
	case y == "":
		return -1
	}

	return strcmp(x, y)
}
