# SPDX-FileCopyrightText: 2024 SUSE LLC
#
# SPDX-License-Identifier: Apache-2.0
ARG GO_VERSION=1.20
FROM mcr.microsoft.com/devcontainers/go:${GO_VERSION}

COPY ./set_container.sh /tmp/scripts/set_container.sh
RUN chmod +x /tmp/scripts/set_container.sh
RUN bash /tmp/scripts/set_container.sh
