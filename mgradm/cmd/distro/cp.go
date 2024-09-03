// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared"
	"github.com/uyuni-project/uyuni-tools/shared/api"
	"github.com/uyuni-project/uyuni-tools/shared/kubernetes"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func umountAndRemove(mountpoint string) {
	umountCmd := []string{
		"/usr/bin/umount",
		mountpoint,
	}

	if err := utils.RunCmd("/usr/bin/sudo", umountCmd...); err != nil {
		log.Error().Err(err).Msgf(L("Unable to unmount ISO image, leaving %s intact"), mountpoint)
	}

	if err := os.Remove(mountpoint); err != nil {
		log.Error().Err(err).Msgf(L("unable to remove temporary directory, leaving %s intact"), mountpoint)
	}
}

func registerDistro(connection *api.ConnectionDetails, distro *types.Distribution, flags *flagpole) error {
	// Fill server FQDN if not provided, ignore error, will be handled later
	if flags.ConnectionDetails.Server == "" {
		flags.ConnectionDetails.Server, _ = getServerFqdn(flags)
		log.Debug().Msgf("Using api-server FQDN '%s'", flags.ConnectionDetails.Server)
	}

	client, err := api.Init(connection)
	if err != nil {
		return utils.Errorf(err, L("unable to login and register the distribution. Manual distro registration is required"))
	}
	data := map[string]interface{}{
		"treeLabel":    distro.TreeLabel,
		"basePath":     distro.BasePath,
		"channelLabel": distro.ChannelLabel,
		"installType":  distro.InstallType,
	}

	_, err = client.Post("kickstart/tree/create", data)
	if err != nil {
		return utils.Errorf(err, L("unable to register the distribution. Manual distro registration is required"))
	}
	log.Info().Msgf(L("Distribution %s successfully registered"), distro.TreeLabel)
	return nil
}

func prepareSource(source string) (string, bool, error) {
	srcdir := source
	needremove := false

	if !utils.FileExists(source) {
		return "", false, fmt.Errorf(L("source %s does not exists"), source)
	}

	if strings.HasSuffix(source, ".iso") {
		log.Debug().Msg("Source is an ISO image")
		tmpdir, err := os.MkdirTemp("", "mgradm-distcp")
		if err != nil {
			return "", needremove, err
		}
		srcdir = tmpdir

		mountCmd := []string{
			"/usr/bin/mount",
			"-o", "ro,loop",
			source,
			srcdir,
		}
		if out, err := utils.RunCmdOutput(zerolog.DebugLevel, "/usr/bin/sudo", mountCmd...); err != nil {
			log.Debug().Msgf("Error mounting ISO image: '%s'", out)
			return "", needremove, fmt.Errorf(L("unable to mount ISO image: %s"), out)
		}
		needremove = true
	}
	return srcdir, needremove, nil
}

func copyDistro(srcdir string, distro *types.Distribution, flags *flagpole) error {
	if len(distro.TreeLabel) == 0 {
		return fmt.Errorf(L("Missing TreeLabel. Please specify distribution name"))
	}

	cnx := shared.NewConnection(flags.Backend, podman.ServerContainerName, kubernetes.ServerFilter)

	const distrosPath = "/srv/www/distributions/"
	dstpath := distrosPath + distro.TreeLabel
	distro.BasePath = dstpath
	if cnx.TestExistenceInPod(dstpath) {
		return fmt.Errorf(L("distribution with same name already exists: %s"), dstpath)
	}

	if _, err := cnx.Exec("sh", "-c", "mkdir -p "+distrosPath); err != nil {
		return utils.Errorf(err, L("cannot create %s path in container"), distrosPath)
	}

	log.Info().Msgf(L("Copying distribution %s"), distro.TreeLabel)
	if err := cnx.Copy(srcdir, "server:"+dstpath, "tomcat", "susemanager"); err != nil {
		return utils.Errorf(err, L("cannot copy %s"), dstpath)
	}
	log.Info().Msgf(L("Distribution has been copied into %s"), distro.BasePath)
	return nil
}

func getServerFqdn(flags *flagpole) (string, error) {
	cnx := shared.NewConnection(flags.Backend, podman.ServerContainerName, kubernetes.ServerFilter)
	fqdn, err := cnx.Exec("sh", "-c", "cat /etc/rhn/rhn.conf 2>/dev/null | grep 'java.hostname' | cut -d' ' -f3")
	return strings.TrimSuffix(string(fqdn), "\n"), err
}

func distroCp(
	globalFlags *types.GlobalFlags,
	flags *flagpole,
	cmd *cobra.Command,
	args []string,
) error {
	source := args[0]
	distroDetails := types.DistributionDetails{}
	if len(args) >= 2 {
		distroDetails.Name = args[1]
		if len(args) > 3 {
			distroDetails.Version = args[2]
			distroDetails.Arch = types.GetArch(args[3])
		}
	}

	attemptRegistration := false
	if flags.ConnectionDetails.User != "" && flags.ConnectionDetails.Password != "" {
		attemptRegistration = true
	}

	srcdir, needremove, err := prepareSource(source)
	if err != nil {
		return err
	}
	if needremove {
		defer umountAndRemove(srcdir)
	}

	distribution := types.Distribution{}
	if err := detectDistro(srcdir, distroDetails, flags, &distribution); err != nil {
		// If we do not want to do the registration, we don't need all the details for mere copy, just name
		if attemptRegistration {
			return err
		}
		log.Debug().Msgf("Would not be able to auto register")
		if len(distroDetails.Name) == 0 {
			// If there is no hint, just use ISO/dir name
			distroDetails.Name = getNameFromSource(source)
		}
		distribution.TreeLabel = distroDetails.Name
	}

	if len(args) == 1 {
		log.Info().Msgf(L("Auto-detected distribution %s"), distribution.TreeLabel)
	}

	if err := copyDistro(srcdir, &distribution, flags); err != nil {
		return err
	}

	if attemptRegistration {
		return registerDistro(&flags.ConnectionDetails, &distribution, flags)
	}

	log.Info().Msgf(L("Continue by registering autoinstallation distribution"))
	return nil
}
