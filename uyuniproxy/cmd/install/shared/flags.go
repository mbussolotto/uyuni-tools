package shared

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	cmd_utils "github.com/uyuni-project/uyuni-tools/uyuniadm/shared/utils"
)

type ProxyInstallFlags struct {
	Image      types.ImageFlags `mapstructure:",squash"`
	MirrorPath string
}

func (flags *ProxyInstallFlags) CheckParameters(cmd *cobra.Command, command string) {

}

func AddInstallFlags(cmd *cobra.Command) {
	cmd_utils.AddImageFlag(cmd)
}
