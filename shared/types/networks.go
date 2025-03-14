// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package types

// PortMap describes a port.
type PortMap struct {
	Service  string
	Name     string
	Exposed  int
	Port     int
	Protocol string
}
