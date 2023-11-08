package shared

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/uyuniadm/shared/utils"
)

type MigrateFlags struct {
	Image types.ImageFlags `mapstructure:",squash"`
}

func AddMigrateFlags(cmd *cobra.Command) {
	utils.AddImageFlag(cmd)
}
