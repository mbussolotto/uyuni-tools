// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package restart

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func podmanRestart(
	globalFlags *types.GlobalFlags,
	flags *restartFlags,
	cmd *cobra.Command,
	args []string,
) error {
	err1 := podman.RestartService(podman.ServerService)
	err2 := podman.RestartInstantiated(podman.ServerAttestationService)
	err3 := podman.RestartInstantiated(podman.HubXmlrpcService)
	return errors.Join(err1, err2, err3)
}
