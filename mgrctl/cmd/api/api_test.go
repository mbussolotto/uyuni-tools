// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func TestNewCommand(t *testing.T) {
	var globalflags types.GlobalFlags
	cmd := NewCommand(&globalflags)
	if cmd == nil {
		t.Error("Unexpected nil command")
	}
}
