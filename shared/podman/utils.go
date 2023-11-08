package podman_utils

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

type PodmanFlags struct {
	Args []string `mapstructure:"arg"`
}

func AddPodmanInstallFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice("podman-arg", []string{}, "Extra arguments to pass to podman")
}

func EnablePodmanSocket() {
	err := utils.RunCmd("systemctl", "enable", "--now", "podman.socket")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to enable podman.socket unit")
	}
}
