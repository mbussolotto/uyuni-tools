// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0
//go:build ptf

package ptf

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/mgrpxy/cmd/support/ptf/kubernetes"
	"github.com/uyuni-project/uyuni-tools/mgrpxy/cmd/support/ptf/podman"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

// NewCommand is the command for creates supportptf.
func NewCommand(globalFlags *types.GlobalFlags) *cobra.Command {
	ptfCmd := &cobra.Command{
		Use:   "ptf",
		Short: L("Install a PTF"),
	}

	ptfCmd.AddCommand(podman.NewCommand(globalFlags))

	if kubernetesCmd := kubernetes.NewCommand(globalFlags); kubernetesCmd != nil {
		ptfCmd.AddCommand(kubernetesCmd)
	}

	return ptfCmd
}
