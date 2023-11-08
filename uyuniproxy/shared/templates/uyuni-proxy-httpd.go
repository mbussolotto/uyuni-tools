package templates

import (
	"io"
	"text/template"
)

const proxyHttpdTemplate = `# uyuni-proxy-httpd.service, generated by uyuniadm
# Use an uyuni-proxy-httpd.service.d/local.conf file to override

[Unit]
Description=Uyuni proxy httpd container service
Wants=network.target
After=network-online.target
BindsTo=uyuni-proxy-pod.service
After=uyuni-proxy-pod.service

[Service]
Environment=PODMAN_SYSTEMD_UNIT=%n
EnvironmentFile={{ .ProxySystemdConfigPath }}
EnvironmentFile={{ .ProxyConfigFolder}}
EnvironmentFile=-{{ .ProxyHttpdConfigPath}}
Restart=on-failure
ExecStartPre=/bin/rm -f %t/uyuni-proxy-httpd.pid %t/uyuni-proxy-httpd.ctr-id

Environment=UYUNI_IMAGE={{ .Image }}
ExecStart=/usr/bin/podman run \
    --conmon-pidfile %t/{{ .NamePrefix }}-proxy-httpd.pid \
    --cidfile %t/{{ .NamePrefix }}-proxy-httpd.ctr-id \
    --cgroups=no-conmon \
    --pod-id-file %t/{{ .NamePrefix }}-proxy-pod.pod-id -d \
    --replace -dt \
	  {{- range $name, $path := .Volumes }}
	  -v {{ $name }}:{{ $path }} \
	  {{- end }}
    --name {{ .NamePrefix }}-proxy-httpd ${UYUNI_IMAGE}

ExecStop=/usr/bin/podman stop --ignore --cidfile %t/{{ .NamePrefix }}-proxy-httpd.ctr-id -t 10
ExecStopPost=/usr/bin/podman rm --ignore -f --cidfile %t/{{ .NamePrefix }}-proxy-httpd.ctr-id
PIDFile=%t/{{ .NamePrefix }}-proxy-httpd.pid
TimeoutStopSec=60
Type=forking

[Install]
WantedBy=multi-user.target default.target
`

type PodmanProxyHttpdTemplateData struct {
	Volumes                map[string]string
	ProxySystemdConfigPath string
	ProxyConfigFolder      string
	ProxyHttpdConfigPath   string
	NamePrefix             string
}

func (data PodmanProxyHttpdTemplateData) Render(wr io.Writer) error {
	t := template.Must(template.New("service").Parse(proxyHttpdTemplate))
	return t.Execute(wr, data)
}
