// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package flagstests

import (
	"testing"

	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/testutils"
)

// PodmanFlagsTestArgs is the values for PodmanFlagsTestArgs.
var PodmanFlagsTestArgs = []string{
	"--podman-arg", "arg1",
	"--podman-arg", "arg2",
}

// AssertPodmanInstallFlags checks that all podman flags are parsed correctly.
func AssertPodmanInstallFlags(t *testing.T, flags *podman.PodmanFlags) {
	testutils.AssertEquals(t, "Error parsing --podman-arg", []string{"arg1", "arg2"}, flags.Args)
}
