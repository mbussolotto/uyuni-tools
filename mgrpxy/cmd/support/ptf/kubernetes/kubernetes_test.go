// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0
//go:build ptf

package kubernetes

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/testutils"
	"github.com/uyuni-project/uyuni-tools/shared/testutils/flagstests"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func TestParamsParsing(t *testing.T) {
	args := []string{}

	args = append(args, flagstests.ImageProxyFlagsTestArgs...)
	args = append(args, flagstests.ProxyHelmFlagsTestArgs...)

	// Test function asserting that the args are properly parsed
	tester := func(_ *types.GlobalFlags, flags *kubernetesPTFFlags,
		_ *cobra.Command, _ []string,
	) error {
		flagstests.AssertProxyImageFlags(t, &flags.UpgradeFlags.ProxyImageFlags)
		flagstests.AssertProxyHelmFlags(t, &flags.UpgradeFlags.Helm)
		return nil
	}

	globalFlags := types.GlobalFlags{}
	cmd := newCmd(&globalFlags, tester)

	testutils.AssertHasAllFlags(t, cmd, args)

	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Errorf("command failed with error: %s", err)
	}
}
