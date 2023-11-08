package podman_utils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

// Ensure the container image is pulled or pull it if the pull policy allows it
func PrepareImage(flags *types.ImageFlags) {
	image := fmt.Sprintf("%s:%s", flags.Name, flags.Tag)
	log.Info().Msgf("Ensure image %s is available", image)

	needsPull := checkImage(image, flags.PullPolicy)
	if needsPull {
		pullImage(image)
	}
}

func checkImage(image string, pullPolicy string) bool {
	if strings.ToLower(pullPolicy) == "always" {
		return true
	}

	out, err := utils.RunCmdOutput(zerolog.DebugLevel, "podman", "images", "--quiet", image)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to check if image %s has already been pulled", image)
	}

	if len(bytes.TrimSpace(out)) == 0 {
		if pullPolicy == "Never" {
			log.Fatal().Msgf("Image %s is not available and cannot be pulled due to policy", image)
		}
		return true
	}
	return false
}

func pullImage(image string) {
	log.Info().Msgf("Running podman pull %s", image)

	err := utils.RunCmdStdMapping("podman", "pull", image)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to pull image")
	}
}
