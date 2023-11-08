package install

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/uyuniproxy/cmd/install/podman"
)

func NewCommand(globalFlags *types.GlobalFlags) *cobra.Command {

	installCmd := &cobra.Command{
		Use:   "install [fqdn]",
		Short: "install a new proxy from scratch",
		Long:  "Install a new proxy from scratch",
	}

	installCmd.AddCommand(podman.NewCommand(globalFlags))

	return installCmd
}
