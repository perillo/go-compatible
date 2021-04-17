// Copyright 2021 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package version

import (
	"testing"
)

// TestParse tests the Parse function and the Version.String method.
func TestParse(t *testing.T) {
	var tests = []struct {
		goversion  string
		version    string
		major      int
		minor      int
		patch      int
		prerelease string
	}{
		{"go1.16", "1.16", 1, 16, 0, ""},
		{"go1.16.1", "1.16.1", 1, 16, 1, ""},
		{"go1.6beta1", "1.6beta1", 1, 6, 0, "beta1"},
		{"go1.17-3f4977bd58", "1.17-3f4977bd58", 1, 17, 0, "-3f4977bd58"},
	}
	for _, test := range tests {
		t.Run(test.goversion, func(t *testing.T) {
			v, err := Parse(test.goversion)
			if err != nil {
				t.Fatalf("expected err == nil, got %q", err)
			}

			if v.Major != test.major {
				t.Errorf("v.Major: got %d, want %d", v.Major, test.major)
			}
			if v.Minor != test.minor {
				t.Errorf("v.Minor: got %d, want %d", v.Minor, test.minor)
			}
			if v.Patch != test.patch {
				t.Errorf("v.Patch: got %d, want %d", v.Patch, test.patch)
			}
			if v.PreRelease != test.prerelease {
				t.Errorf("v.PreRelease: got %s, want %s", v.PreRelease,
					test.prerelease)
			}

			if s := v.String(); s != test.version {
				t.Errorf("v.String(): got %q, want %q", s, test.version)
			}
		})
	}
}
