// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package podman

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/mgradm/shared/podman"
	shared_podman "github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func upgradePodman(_ *types.GlobalFlags, flags *podmanUpgradeFlags, _ *cobra.Command, _ []string) error {
	hostData, err := shared_podman.InspectHost()
	if err != nil {
		return err
	}

	authFile, cleaner, err := shared_podman.PodmanLogin(hostData, flags.Image.Registry, flags.SCC)
	if err != nil {
		return err
	}
	defer cleaner()

	return podman.Upgrade(
		authFile, flags.Image.Registry, flags.Image, flags.DBUpgradeImage, flags.Coco, flags.HubXmlrpc,
	)
}
