// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package types

// GlobalFlags represents the flags used by all commands.
type GlobalFlags struct {
	Registry   string //set using local flag registry (from image group)
	ConfigPath string
	LogLevel   string
}
