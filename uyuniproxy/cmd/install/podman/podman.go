package podman

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	podman_utils "github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
	"github.com/uyuni-project/uyuni-tools/uyuniproxy/cmd/install/shared"
)

type podmanProxyInstallFlags struct {
	shared.ProxyInstallFlags `mapstructure:",squash"`
	Podman                   podman_utils.PodmanFlags
}

func NewCommand(globalFlags *types.GlobalFlags) *cobra.Command {

	podmanCmd := &cobra.Command{
		Use:   "podman [fqdn]",
		Short: "install a new proxy on podman from scratch",
		Long: `Install a new proxy on podman from scratch

The install podman command assumes podman is installed locally

NOTE: for now installing on a remote podman is not supported!
`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			viper := utils.ReadConfig(globalFlags.ConfigPath, "admconfig", cmd)
			var flags podmanProxyInstallFlags
			if err := viper.Unmarshal(&flags); err != nil {
				log.Fatal().Err(err).Msgf("Failed to unmarshall configuration")
			}
			flags.CheckParameters(cmd, "podman")
			installForPodman(globalFlags, &flags, cmd, args)
		},
	}

	shared.AddInstallFlags(podmanCmd)
	podman_utils.AddPodmanInstallFlag(podmanCmd)

	return podmanCmd
}
