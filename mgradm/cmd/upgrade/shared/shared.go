// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/uyuni-project/uyuni-tools/shared"
)

// CompareVersion compare the server image version and the server deployed  version.
func CompareVersion(imageVersion string, deployedVersion string) int {
	imageVersionCleaned := strings.ReplaceAll(imageVersion, ".", "")
	imageVersionCleaned = strings.TrimSpace(imageVersionCleaned)
	imageVersionInt, _ := strconv.Atoi(imageVersionCleaned)
	deployedVersionCleaned := strings.ReplaceAll(deployedVersion, ".", "")
	deployedVersionInt, _ := strconv.Atoi(deployedVersionCleaned)
	return imageVersionInt - deployedVersionInt
}

// SanityCheck verifies if an upgrade can be run.
func SanityCheck(cnx *shared.Connection, inspectedValues map[string]string, serverImage string) error {
	cnx_args := []string{"s/Uyuni release //g", "/etc/uyuni-release"}
	current_uyuni_release, err := cnx.Exec("sed", cnx_args...)
	if err != nil {
		return fmt.Errorf("failed to read current uyuni release: %s", err)
	}
	log.Debug().Msgf("Current release is %s", string(current_uyuni_release))
	if (len(inspectedValues["uyuni_release"])) <= 0 {
		return fmt.Errorf("cannot fetch release from image %s", serverImage)
	}
	log.Debug().Msgf("Image %s is %s", serverImage, inspectedValues["uyuni_release"])
	if CompareVersion(inspectedValues["uyuni_release"], string(current_uyuni_release)) <= 0 {
		return fmt.Errorf("cannot downgrade from version %s to %s", string(current_uyuni_release), inspectedValues["uyuni_release"])
	}
	if (len(inspectedValues["image_pg_version"])) <= 0 {
		return fmt.Errorf("cannot fetch postgresql version from %s", serverImage)
	}
	log.Debug().Msgf("Image %s has PostgreSQL %s", serverImage, inspectedValues["image_pg_version"])
	if (len(inspectedValues["current_pg_version"])) <= 0 {
		return fmt.Errorf("posgresql is not installed in the current deployment")
	}
	log.Debug().Msgf("Current deployment has PostgreSQL %s", inspectedValues["current_pg_version"])

	return nil
}