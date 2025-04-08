// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

//go:build nok8s

package restart

import (
	"errors"

	"github.com/spf13/cobra"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func kubernetesRestart(
	_ *types.GlobalFlags,
	_ *restartFlags,
	_ *cobra.Command,
	_ []string,
) error {
	return errors.New(L("built without kubernetes support"))
}
