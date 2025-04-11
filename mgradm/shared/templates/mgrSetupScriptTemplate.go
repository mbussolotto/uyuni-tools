// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"io"
	"text/template"
)

//nolint:lll
const mgrSetupScriptTemplate = `#!/bin/sh
if test -e /root/.MANAGER_SETUP_COMPLETE; then
	echo "Server appears to be already configured. Installation options may be ignored."
	exit 0
fi

{{- range $name, $value := .Env }}
export {{ $name }}='{{ $value }}'
{{- end }}

{{- if .DebugJava }}
echo 'JAVA_OPTS=" $JAVA_OPTS -Xdebug -Xrunjdwp:transport=dt_socket,address=*:8003,server=y,suspend=n" ' >> /etc/tomcat/conf.d/remote_debug.conf
echo 'JAVA_OPTS=" $JAVA_OPTS -Xdebug -Xrunjdwp:transport=dt_socket,address=*:8001,server=y,suspend=n" ' >> /etc/rhn/taskomatic.conf
echo 'JAVA_OPTS=" $JAVA_OPTS -Xdebug -Xrunjdwp:transport=dt_socket,address=*:8002,server=y,suspend=n" ' >> /usr/share/rhn/config-defaults/rhn_search_daemon.conf
{{- end }}

/usr/lib/susemanager/bin/mgr-setup -s -n
RESULT=$?

# clean before leaving
rm $0
exit $RESULT
`

// MgrSetupScriptTemplateData represents information used to create setup script.
type MgrSetupScriptTemplateData struct {
	Env       map[string]string
	DebugJava bool
}

// Render will create setup script.
func (data MgrSetupScriptTemplateData) Render(wr io.Writer) error {
	t := template.Must(template.New("script").Parse(mgrSetupScriptTemplate))
	return t.Execute(wr, data)
}
