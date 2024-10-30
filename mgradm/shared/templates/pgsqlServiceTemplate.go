// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"io"
	"text/template"

	"github.com/uyuni-project/uyuni-tools/shared/types"
)

const pgsqlServiceTemplate = `# uyuni-pgsql-server.service, generated by mgradm
# Use an uyuni-pgsql-server.service.d/local.conf file to override

[Unit]
Description=Uyuni postgres server image container service
Wants=network.target
After=network-online.target
RequiresMountsFor=%t/containers

[Service]
Environment=PODMAN_SYSTEMD_UNIT=%n
Restart=on-failure
ExecStartPre=/bin/rm -f %t/uyuni-pgsql-server.pid %t/%n.ctr-id
ExecStartPre=/usr/bin/podman rm --ignore --force -t 10 {{ .NamePrefix }}-pgsql-server
ExecStart=/bin/sh -c '/usr/bin/podman run \
	--conmon-pidfile %t/uyuni-pgsql-server.pid \
	--cidfile=%t/%n.ctr-id \
	--cgroups=no-conmon \
	--shm-size=0 \
	--shm-size-systemd=0 \
	--sdnotify=conmon \
	-d \
	--name {{ .NamePrefix }}-pgsql-server \
	--hostname {{ .NamePrefix }}-pgsql-server.mgr.internal \
	{{ .Args }} \
	{{- range .Ports }}
	-p {{ .Exposed }}:{{ .Port }}{{if .Protocol}}/{{ .Protocol }}{{end}} \
        {{- if $.IPV6Enabled }}
	-p [::]:{{ .Exposed }}:{{ .Port }}{{if .Protocol}}/{{ .Protocol }}{{end}} \
        {{- end }}
	{{- end }}
        {{- range .Volumes }}
        -v {{ .Name }}:{{ .MountPath }}:z \
        {{- end }}
	-e TZ=${TZ} \
	-e POSTGRES_PASSWORD \
	--network {{ .Network }} \
	${PODMAN_EXTRA_ARGS} ${UYUNI_IMAGE}'
ExecStop=/usr/bin/podman stop \
	--ignore -t 10 \
	--cidfile=%t/%n.ctr-id
ExecStopPost=/usr/bin/podman rm \
	-f \
	--ignore -t 10 \
	--cidfile=%t/%n.ctr-id

PIDFile=%t/uyuni-pgsql-server.pid
TimeoutStopSec=180
TimeoutStartSec=900
Type=forking

[Install]
WantedBy=multi-user.target default.target
`

// PostgresServiceTemplateData POD information to create systemd file.
type PgsqlServiceTemplateData struct {
	Volumes     []types.VolumeMount
	NamePrefix  string
	Args        string
	Ports       []types.PortMap
	Image       string
	Network     string
	IPV6Enabled bool
}

// Render will create the systemd configuration file.
func (data PgsqlServiceTemplateData) Render(wr io.Writer) error {
	t := template.Must(template.New("service").Parse(pgsqlServiceTemplate))
	return t.Execute(wr, data)
}
