// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package podman

import (
	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func StartServices() error {
	return utils.JoinErrors(
		podman.StartInstantiated(podman.ServerAttestationService),
		podman.StartInstantiated(podman.HubXmlrpcService),
		podman.StartService(podman.ServerService),
	)
}

func StopServices() error {
	return utils.JoinErrors(
		podman.StopInstantiated(podman.ServerAttestationService),
		podman.StopInstantiated(podman.HubXmlrpcService),
		podman.StopService(podman.ServerService),
	)
}
