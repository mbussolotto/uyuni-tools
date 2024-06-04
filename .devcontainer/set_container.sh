#!/usr/bin/sh

# SPDX-FileCopyrightText: 2024 SUSE LLC
#
# SPDX-License-Identifier: Apache-2.0

go mod init temp-module
go get github.com/uyuni-project/uyuni-tools
go mod download
