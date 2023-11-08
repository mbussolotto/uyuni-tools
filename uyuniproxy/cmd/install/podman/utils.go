package podman

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	podman_utils "github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
	"github.com/uyuni-project/uyuni-tools/uyuniproxy/shared/podman"
)

func waitForSystemStart(cnx *utils.Connection, flags *podmanProxyInstallFlags) {
	// Setup the systemd service configuration options
	image := fmt.Sprintf("%s:%s", flags.Image.Name, flags.Image.Tag)

	podmanArgs := flags.Podman.Args
	if flags.MirrorPath != "" {
		podmanArgs = append(podmanArgs, "-v", flags.MirrorPath+":/mirror")
	}

	podman.GenerateSystemdService(image, podmanArgs)

	log.Info().Msg("Waiting for the proxy to start...")
	// Start the service

	if err := utils.RunCmd("systemctl", "enable", "--now", "uyuni-proxy"); err != nil {
		log.Fatal().Err(err).Msg("Failed to enable uyuni-proxy systemd service")
	}

	cnx.WaitForServer()
}

func installForPodman(globalFlags *types.GlobalFlags, flags *podmanProxyInstallFlags, cmd *cobra.Command, args []string) {
	fqdn := getFqdn(args)
	log.Info().Msgf("setting up proxy with the FQDN '%s'", fqdn)

	podman_utils.PrepareImage(&flags.Image)

	cnx := utils.NewConnection("podman")
	waitForSystemStart(cnx, flags)

	log.Info().Msg("run setup command in the container")

	podman_utils.EnablePodmanSocket()
}

func getFqdn(args []string) string {
	if len(args) == 1 {
		return args[0]
	} else {
		fqdn_b, err := utils.RunCmdOutput(zerolog.DebugLevel, "hostname", "-f")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to compute proxy FQDN")
		}
		return string(fqdn_b)
	}
}
