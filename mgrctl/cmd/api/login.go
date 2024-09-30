// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/uyuni-project/uyuni-tools/shared/api"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func runLogin(globalFlags *types.GlobalFlags, flags *apiFlags, cmd *cobra.Command, args []string) error {
	log.Debug().Msg("Running login command")

	if api.IsAlreadyLoggedIn() && !flags.ForceLogin {
		return fmt.Errorf(L("Refusing to overwrite existing login. Use --force to ignore this check."))
	}

	utils.AskIfMissing(&flags.Server, cmd.Flag("api-server").Usage, 0, 0, utils.IsWellFormedFQDN)
	utils.AskIfMissing(&flags.User, cmd.Flag("api-user").Usage, 0, 0, nil)
	utils.AskPasswordIfMissingOnce(&flags.Password, cmd.Flag("api-password").Usage, 0, 0)

	client, err := api.Init(&flags.ConnectionDetails)
	if err != nil {
		return err
	}
	if err := client.Login(); err != nil {
		return utils.Errorf(err, L("Failed to validate credentials."))
	}
	if err := api.StoreLoginCreds(client); err != nil {
		return err
	}

	log.Info().Msg(L("Login credentials verified."))
	return nil
}

func runLogout(globalFlags *types.GlobalFlags, flags *apiFlags, cmd *cobra.Command, args []string) error {
	log.Debug().Msg("Running logout command")

	if err := api.RemoveLoginCreds(); err != nil {
		return err
	}
	log.Info().Msg(L("Successfully logged out"))
	return nil
}
