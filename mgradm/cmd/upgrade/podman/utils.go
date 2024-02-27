// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package podman

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/mgradm/cmd/inspect"
	upgrade_shared "github.com/uyuni-project/uyuni-tools/mgradm/cmd/upgrade/shared"

	"github.com/uyuni-project/uyuni-tools/mgradm/shared/podman"

	adm_utils "github.com/uyuni-project/uyuni-tools/mgradm/shared/utils"
	"github.com/uyuni-project/uyuni-tools/shared"
	shared_podman "github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func upgradePodman(globalFlags *types.GlobalFlags, flags *podmanUpgradeFlags, cmd *cobra.Command, args []string) error {
	serverImage, err := utils.ComputeImage(flags.Image.Name, flags.Image.Tag)
	if err != nil {
		return fmt.Errorf("failed to compute image URL")
	}

	inspectedValues, err := inspect.InspectPodman(serverImage, flags.Image.PullPolicy)
	if err != nil {
		return fmt.Errorf("cannot inspect podman values: %s", err)
	}

	cnx := shared.NewConnection("podman", shared_podman.ServerContainerName, "")

	err = upgrade_shared.SanityCheck(cnx, inspectedValues, serverImage)
	if err != nil {
		return err
	}

	scriptDir, err := os.MkdirTemp("", "mgradm-*")
	defer os.RemoveAll(scriptDir)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory")
	}

	err = shared_podman.StopService(shared_podman.ServerService)
	if err != nil {
		return fmt.Errorf("cannot stop service %s", err)
	}

	defer func() {
		err = shared_podman.StartService(shared_podman.ServerService)
	}()

	if inspectedValues["image_pg_version"] > inspectedValues["current_pg_version"] {
		log.Info().Msgf("Previous postgresql is %s, instead new one is %s. Performing a DB migration...", inspectedValues["current_pg_version"], inspectedValues["image_pg_version"])
		extraArgs := []string{
			"-v", scriptDir + ":/var/lib/uyuni-tools/",
			"--security-opt", "label:disable",
		}

		migrationImageUrl := ""
		if flags.MigrationImage.Name == "" {
			migrationImageUrl, err = utils.ComputeImage(flags.Image.Name, flags.Image.Tag, fmt.Sprintf("-migration-%s-%s", inspectedValues["current_pg_version"], inspectedValues["image_pg_version"]))
			if err != nil {
				return fmt.Errorf("failed to compute image URL %s", err)
			}
		} else {
			migrationImageUrl, err = utils.ComputeImage(flags.MigrationImage.Name, flags.Image.Tag)
			if err != nil {
				return fmt.Errorf("failed to compute image URL %s", err)
			}
		}

		err = shared_podman.PrepareImage(migrationImageUrl, flags.Image.PullPolicy)
		if err != nil {
			return err
		}

		log.Info().Msgf("Using migration image %s", migrationImageUrl)

		scriptName, err := adm_utils.GeneratePgsqlVersionUpgradeScript(scriptDir, inspectedValues["current_pg_version"], inspectedValues["image_pg_version"], false)
		if err != nil {
			return fmt.Errorf("cannot generate postgresql database version upgrade script %s", err)
		}

		err = podman.RunContainer("uyuni-upgrade-pgsql", migrationImageUrl, extraArgs,
			[]string{"/var/lib/uyuni-tools/" + scriptName})
		if err != nil {
			return err
		}
	} else if inspectedValues["image_pg_version"] == inspectedValues["current_pg_version"] {
		log.Info().Msgf("Upgrading to %s without changing PostgreSQL version", inspectedValues["uyuni_release"])
	} else {
		return fmt.Errorf("trying to downgrade postgresql from %s to %s", inspectedValues["current_pg_version"], inspectedValues["image_pg_version"])
	}

	extraArgs := []string{
		"-v", scriptDir + ":/var/lib/uyuni-tools/",
		"--security-opt", "label:disable",
	}

	scriptName, err := adm_utils.GenerateFinalizePostgresMigrationScript(scriptDir, true, inspectedValues["current_pg_version"] != inspectedValues["image_pg_version"], true, true, false)
	if err != nil {
		return fmt.Errorf("cannot generate postgresql migration finalization script %s", err)
	}
	err = podman.RunContainer("uyuni-finalize-pgsql", serverImage, extraArgs,
		[]string{"/var/lib/uyuni-tools/" + scriptName})
	if err != nil {
		return err
	}

	err = shared_podman.GenerateSystemdConfFile("uyuni-server", "Service", "Environment=UYUNI_IMAGE="+serverImage)
	if err != nil {
		return err
	}
	log.Info().Msg("Waiting for the server to start...")
	return shared_podman.ReloadDaemon(false)
}
