// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package podman

import (
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

const (
	// DBUserSecret is the name of the podman secret containing the database username.
	DBUserSecret = "uyuni-db-user"
	// DBPassSecret is the name of the podman secret containing the database password.
	DBPassSecret = "uyuni-db-pass"
	// ReportDBUserSecret is the name of the podman secret containing the report database username.
	ReportDBUserSecret = "uyuni-reportdb-user"
	// ReportDBPassSecret is the name of the podman secret containing the report database password.
	ReportDBPassSecret = "uyuni-reportdb-pass"
	// DBUserSecret is the name of the podman secret containing the database admin username.
	DBAdminUserSecret = "uyuni-db-admin-user"
	// DBAdminPassSecret is the name of the podman secret containing the database admin password.
	DBAdminPassSecret = "uyuni-db-admin-pass"
	// CASecret is the name of the podman secret containing the CA certificate.
	CASecret = "uyuni-ca"
	// DBSSLCertSecret is the name of the podman secret containing the report database certificate.
	DBSSLCertSecret = "uyuni-db-cert"
	// DBSSLKeySecret is the name of the podman secret containing the report database SSL certificate key.
	DBSSLKeySecret = "uyuni-db-key"
)

// CreateCredentialsSecrets creates the podman secrets, one for the user name and one for the password.
func CreateCredentialsSecrets(userSecret string, user string, passwordSecret string, password string) error {
	if err := createSecret(userSecret, user); err != nil {
		return err
	}
	return createSecret(passwordSecret, password)
}

// CreateDBTLSSecrets creates the SSL CA, Certificate and key secrets.
func CreateDBTLSSecrets(caPath string, certPath string, keyPath string) error {
	if err := createSecretFromFile(CASecret, caPath); err != nil {
		return utils.Errorf(err, L("failed to create %s secret"), CASecret)
	}

	if err := createSecretFromFile(DBSSLCertSecret, certPath); err != nil {
		return utils.Errorf(err, L("failed to create %s secret"), DBSSLCertSecret)
	}

	if err := createSecretFromFile(DBSSLKeySecret, keyPath); err != nil {
		return utils.Errorf(err, L("failed to create %s secret"), DBSSLKeySecret)
	}

	return nil
}

// createSecret creates a podman secret.
func createSecret(name string, value string) error {
	tmpDir, cleaner, err := utils.TempDir()
	if err != nil {
		return err
	}
	defer cleaner()

	secretFile := path.Join(tmpDir, "secret")
	if err := os.WriteFile(secretFile, []byte(value), 0600); err != nil {
		return utils.Errorf(err, L("failed to write %s secret to file"), name)
	}

	return createSecretFromFile(name, secretFile)
}

// createSecretFromFile creates a podman secret from a file.
func createSecretFromFile(name string, secretFile string) error {
	if hasSecret(name) {
		return nil
	}

	if err := utils.RunCmd("podman", "secret", "create", name, secretFile); err != nil {
		return utils.Errorf(err, L("failed to create podman secret %s"), name)
	}

	return nil
}

func hasSecret(name string) bool {
	return utils.RunCmd("podman", "secret", "exists", name) == nil
}

// DeleteSecret removes a podman secret.
func DeleteSecret(name string, dryRun bool) {
	if !hasSecret(name) {
		return
	}

	args := []string{"secret", "rm", name}
	command := "podman " + strings.Join(args, " ")
	if dryRun {
		log.Info().Msgf(L("Would run %s"), command)
	} else {
		log.Info().Msgf(L("Run %s"), command)
		if err := utils.RunCmd("podman", args...); err != nil {
			log.Error().Err(err).Msgf(L("Failed to delete %s secret"), name)
		}
	}
}
