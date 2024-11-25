// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"path"

	"github.com/uyuni-project/uyuni-tools/shared/types"
)

// ServerInspector inspects a running server container or its image.
type ServerInspector struct {
	BaseInspector
}

// NewServerInspector creates a new ServerInspector generating the inspection script and data in scriptDir.
func NewServerInspector(scriptDir string) ServerInspector {
	base := BaseInspector{
		Values: []types.InspectData{
			types.NewInspectData(
				"uyuni_release",
				"cat /etc/*release | grep 'Uyuni release' | cut -d ' ' -f3 || true"),
			types.NewInspectData(
				"suse_manager_release",
				"sed 's/.*(\\([0-9.]\\+\\).*/\\1/g' /etc/susemanager-release || true"),
			types.NewInspectData(
				"fqdn",
				"cat /etc/rhn/rhn.conf 2>/dev/null | grep -m1 '^java.hostname' | cut -d' ' -f3 || true"),
			types.NewInspectData(
				"image_pg_version",
				"rpm -qa --qf '%{VERSION}\\n' 'name=postgresql[0-8][0-9]-server'  | cut -d. -f1 | sort -n | tail -1 || true"),
			types.NewInspectData("current_pg_version",
				"(test -e /var/lib/pgsql/data/PG_VERSION && cat /var/lib/pgsql/data/PG_VERSION) || true"),
			types.NewInspectData("current_pg_version_not_migrated",
				"(test -e /var/lib/pgsql/data/data/PG_VERSION && cat /var/lib/pgsql/data/data/PG_VERSION) || true"),
			types.NewInspectData("db_user",
				"cat /etc/rhn/rhn.conf 2>/dev/null | grep -m1 '^db_user' | cut -d' ' -f3 || true"),
			types.NewInspectData("db_password",
				"cat /etc/rhn/rhn.conf 2>/dev/null | grep -m1 '^db_password' | cut -d' ' -f3 || true"),
			types.NewInspectData("db_name",
				"cat /etc/rhn/rhn.conf 2>/dev/null | grep -m1 '^db_name' | cut -d' ' -f3 || true"),
			types.NewInspectData("db_port",
				"cat /etc/rhn/rhn.conf 2>/dev/null | grep -m1 '^db_port' | cut -d' ' -f3 || true"),
			types.NewInspectData("db_host",
				"cat /etc/rhn/rhn.conf 2>/dev/null | grep -m1 '^db_host' | cut -d' ' -f3 || true"),
			types.NewInspectData("report_db_host",
				"cat /etc/rhn/rhn.conf 2>/dev/null | grep -m1 '^report_db_host' | cut -d' ' -f3 || true"),
		},
		ScriptDir: scriptDir,
		DataPath:  path.Join(InspectContainerDirectory, inspectDataFile),
	}
	return ServerInspector{
		BaseInspector: base,
	}
}

// CommonInspectData are data common between the migration source inspect and server inspector results.
type CommonInspectData struct {
	CurrentPgVersion            string `mapstructure:"current_pg_version"`
	CurrentPgVersionNotMigrated string `mapstructure:"current_pg_version_not_migrated"`
	ImagePgVersion              string `mapstructure:"image_pg_version"`
	DBUser                      string `mapstructure:"db_user"`
	DBPassword                  string `mapstructure:"db_password"`
	DBName                      string `mapstructure:"db_name"`
	DBPort                      int    `mapstructure:"db_port"`
	DBHost                      string `mapstructure:"db_host"`
	ReportDBHost                string `mapstructure:"report_db_host"`
}

// ServerInspectData are the data extracted by a server inspector.
type ServerInspectData struct {
	CommonInspectData  `mapstructure:",squash"`
	UyuniRelease       string `mapstructure:"uyuni_release"`
	SuseManagerRelease string `mapstructure:"suse_manager_release"`
	Fqdn               string
}

// ReadInspectData parses the data generated by the server inspector.
func (i *ServerInspector) ReadInspectData() (*ServerInspectData, error) {
	return ReadInspectData[ServerInspectData](path.Join(i.ScriptDir, inspectDataFile))
}
