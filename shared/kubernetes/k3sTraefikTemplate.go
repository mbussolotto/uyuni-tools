// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"io"
	"text/template"

	"github.com/uyuni-project/uyuni-tools/shared/types"
)

const k3sTraefikConfigTemplate = `apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: traefik
  namespace: kube-system
spec:
  valuesContent: |-
    ports:
{{- range .Ports }}
      {{ .Name }}:
        port: {{ .Port }}
        {{- if $.ExposeBoolean }}
        expose: true
        {{- else }}
        expose:
          default: true
        {{- end }}
        exposedPort: {{ .Exposed }}
        {{- if eq .Protocol "udp" }}
        protocol: UDP
        {{- else }}
        protocol: TCP
        {{- end }}
{{- end }}
`

// K3sTraefikConfigTemplateData represents information used to create K3s Traefik helm chart.
type K3sTraefikConfigTemplateData struct {
	Ports []types.PortMap
	// Set to true before traefik chart v27
	ExposeBoolean bool
}

// Render will create the helm chart configuation for K3sTraefik.
func (data K3sTraefikConfigTemplateData) Render(wr io.Writer) error {
	t := template.Must(template.New("k3sTraefikConfig").Parse(k3sTraefikConfigTemplate))
	return t.Execute(wr, data)
}
