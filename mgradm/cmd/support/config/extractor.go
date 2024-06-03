// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared"
	"github.com/uyuni-project/uyuni-tools/shared/kubernetes"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func extract(globalFlags *types.GlobalFlags, flags *configFlags, cmd *cobra.Command, args []string) error {
	cnx := shared.NewConnection(flags.Backend, podman.ServerContainerName, kubernetes.ServerFilter)
	supportConfigFiles, err := shared.RunSupportConfig(cnx)
	if err != nil {
		return err
	}

	// Copy the generated file locally
	tmpDir, err := os.MkdirTemp("", "mgradm-*")
	if err != nil {
		return utils.Errorf(err, L("failed to create temporary directory"))
	}

	defer os.RemoveAll(tmpDir)
	hostSupportConfigFiles, err := utils.RunSupportConfigOnHost(tmpDir)
	if err != nil {
		return err
	}

	supportConfigFiles = append(supportConfigFiles, hostSupportConfigFiles...)

	// TODO Get cluster infos in case of kubernetes

	if err := utils.CreateSupportConfigTarball(flags.Output, supportConfigFiles); err != nil {
		return err
	}

	return nil
}
