// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package stop

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func podmanStop(
	globalFlags *types.GlobalFlags,
	flags *stopFlags,
	cmd *cobra.Command,
	args []string,
) error {
	if podman.HasService(podman.ServerAttestationService) {
		if err := podman.StopService(podman.ServerAttestationService); err != nil {
			return err
		}
	}
	return podman.StopService(podman.ServerService)
}
