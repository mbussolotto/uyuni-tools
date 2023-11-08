package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
	"github.com/uyuni-project/uyuni-tools/uyuniproxy/cmd/install"
)

// NewCommand returns a new cobra.Command implementing the root command for kinder
func NewUyuniproxyCommand() *cobra.Command {
	utils.LogInit("uyuniproxy", true)
	globalFlags := &types.GlobalFlags{}
	rootCmd := &cobra.Command{
		Use:     "uyuniproxy",
		Short:   "Uyuni proxy tool",
		Long:    "Uyuni proxy tool used to help user administer uyuni proxies on kubernetes and podman",
		Version: "0.0.1",
	}

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		utils.SetLogLevel(globalFlags.LogLevel)
		log.Info().Msgf("Executing command: %s", cmd.Name())
	}

	rootCmd.PersistentFlags().StringVarP(&globalFlags.ConfigPath, "config", "c", "", "configuration file path")
	rootCmd.PersistentFlags().StringVar(&globalFlags.LogLevel, "logLevel", "", "application log level (trace|debug|info|warn|error|fatal|panic)")

	installCmd := install.NewCommand(globalFlags)
	rootCmd.AddCommand(installCmd)

	return rootCmd
}
