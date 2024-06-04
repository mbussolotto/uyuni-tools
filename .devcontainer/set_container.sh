#!/usr/bin/sh

# SPDX-FileCopyrightText: 2024 SUSE LLC
#
# SPDX-License-Identifier: Apache-2.0

wget https://raw.githubusercontent.com/uyuni-project/uyuni-tools/main/go.mod
wget https://raw.githubusercontent.com/uyuni-project/uyuni-tools/main/go.sum
go mod download
rm go.mod
rm go.sum
