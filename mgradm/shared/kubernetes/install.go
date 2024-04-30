// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/mgradm/shared/ssl"
	cmd_utils "github.com/uyuni-project/uyuni-tools/mgradm/shared/utils"
	"github.com/uyuni-project/uyuni-tools/shared"
	"github.com/uyuni-project/uyuni-tools/shared/kubernetes"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

// HELM_APP_NAME is the Helm application name.
const HELM_APP_NAME = "uyuni"

// Deploy execute a deploy of a given image and helm to a cluster.
func Deploy(cnx *shared.Connection, imageFlags *types.ImageFlags,
	helmFlags *cmd_utils.HelmFlags, sslFlags *cmd_utils.SslCertFlags, clusterInfos *kubernetes.ClusterInfos,
	fqdn string, debug bool, helmArgs ...string) error {
	// If installing on k3s, install the traefik helm config in manifests
	isK3s := clusterInfos.IsK3s()
	IsRke2 := clusterInfos.IsRke2()
	if isK3s {
		InstallK3sTraefikConfig(debug)
	} else if IsRke2 {
		kubernetes.InstallRke2NginxConfig(utils.TCP_PORTS, utils.UDP_PORTS, helmFlags.Uyuni.Namespace)
	}

	serverImage, err := utils.ComputeImage(imageFlags.Name, imageFlags.Tag)
	if err != nil {
		return fmt.Errorf(L("failed to compute image URL: %s"), err)
	}

	// Install the uyuni server helm chart
	err = UyuniUpgrade(serverImage, imageFlags.PullPolicy, helmFlags, clusterInfos.GetKubeconfig(), fqdn, clusterInfos.Ingress, helmArgs...)
	if err != nil {
		return fmt.Errorf(L("cannot upgrade: %s"), err)
	}

	// Wait for the pod to be started
	err = kubernetes.WaitForDeployment(helmFlags.Uyuni.Namespace, HELM_APP_NAME, "uyuni")
	if err != nil {
		return fmt.Errorf(L("cannot deploy: %s"), err)
	}
	return cnx.WaitForServer()
}

// DeployCertificate executre a deploy a new certificate given an helm.
func DeployCertificate(helmFlags *cmd_utils.HelmFlags, sslFlags *cmd_utils.SslCertFlags, rootCa string,
	ca *ssl.SslPair, kubeconfig string, fqdn string, imagePullPolicy string) ([]string, error) {
	helmArgs := []string{}
	if sslFlags.UseExisting() {
		DeployExistingCertificate(helmFlags, sslFlags, kubeconfig)
	} else {
		// Install cert-manager and a self-signed issuer ready for use
		issuerArgs, err := installSslIssuers(helmFlags, sslFlags, rootCa, ca, kubeconfig, fqdn, imagePullPolicy)
		if err != nil {
			return []string{}, fmt.Errorf(L("cannot install cert-manager and self-sign issuer: %s"), err)
		}
		helmArgs = append(helmArgs, issuerArgs...)

		// Extract the CA cert into uyuni-ca config map as the container shouldn't have the CA secret
		extractCaCertToConfig()
	}

	return helmArgs, nil
}

// DeployExistingCertificate execute a deploy of an existing certificate.
func DeployExistingCertificate(helmFlags *cmd_utils.HelmFlags, sslFlags *cmd_utils.SslCertFlags, kubeconfig string) {
	// Deploy the SSL Certificate secret and CA configmap
	serverCrt, rootCaCrt := ssl.OrderCas(&sslFlags.Ca, &sslFlags.Server)
	serverKey := utils.ReadFile(sslFlags.Server.Key)
	installTlsSecret(helmFlags.Uyuni.Namespace, serverCrt, serverKey, rootCaCrt)

	// Extract the CA cert into uyuni-ca config map as the container shouldn't have the CA secret
	extractCaCertToConfig()
}

// UyuniUpgrade runs an helm upgrade using images and helm configuration as parameters.
func UyuniUpgrade(serverImage string, pullPolicy string, helmFlags *cmd_utils.HelmFlags, kubeconfig string,
	fqdn string, ingress string, helmArgs ...string) error {
	log.Info().Msg(L("Installing Uyuni"))

	// The guessed ingress is passed before the user's value to let the user override it in case we got it wrong.
	helmParams := []string{
		"--set", "ingress=" + ingress,
	}

	extraValues := helmFlags.Uyuni.Values
	if extraValues != "" {
		helmParams = append(helmParams, "-f", extraValues)
	}

	// The values computed from the command line need to be last to override what could be in the extras
	helmParams = append(helmParams,
		"--set", "images.server="+serverImage,
		"--set", "pullPolicy="+kubernetes.GetPullPolicy(pullPolicy),
		"--set", "fqdn="+fqdn)

	helmParams = append(helmParams, helmArgs...)

	namespace := helmFlags.Uyuni.Namespace
	chart := helmFlags.Uyuni.Chart
	version := helmFlags.Uyuni.Version
	return kubernetes.HelmUpgrade(kubeconfig, namespace, true, "", HELM_APP_NAME, chart, version, helmParams...)
}

// Upgrade will upgrade a server in a kubernetes cluster.
func Upgrade(
	globalFlags *types.GlobalFlags,
	image *types.ImageFlags,
	migrationImage *types.ImageFlags,
	helm cmd_utils.HelmFlags,
	cmd *cobra.Command,
	args []string,
) error {
	for _, binary := range []string{"kubectl", "helm"} {
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf(L("install %s before running this command"), binary)
		}
	}
	cnx := shared.NewConnection("kubectl", "", kubernetes.ServerFilter)

	serverImage, err := utils.ComputeImage(image.Name, image.Tag)
	if err != nil {
		return fmt.Errorf(L("failed to compute image URL: %s"), err)
	}

	inspectedValues, err := kubernetes.InspectKubernetes(serverImage, image.PullPolicy)
	if err != nil {
		return fmt.Errorf(L("cannot inspect kubernetes values: %s"), err)
	}

	err = cmd_utils.SanityCheck(cnx, inspectedValues, serverImage)
	if err != nil {
		return err
	}

	fqdn, exist := inspectedValues["fqdn"]
	if !exist {
		return fmt.Errorf(L("inspect function did non return fqdn value"))
	}

	clusterInfos, err := kubernetes.CheckCluster()
	if err != nil {
		return err
	}
	kubeconfig := clusterInfos.GetKubeconfig()

	scriptDir, err := os.MkdirTemp("", "mgradm-*")
	defer os.RemoveAll(scriptDir)
	if err != nil {
		return fmt.Errorf(L("failed to create temporary directory: %s"), err)
	}

	//this is needed because folder with script needs to be mounted
	//check the node before scaling down
	nodeName, err := kubernetes.GetNode("uyuni")
	if err != nil {
		return fmt.Errorf(L("cannot find node running uyuni: %s"), err)
	}

	err = kubernetes.ReplicasTo(kubernetes.ServerFilter, 0)
	if err != nil {
		return fmt.Errorf(L("cannot set replica to 0: %s"), err)
	}

	defer func() {
		// if something is running, we don't need to set replicas to 1
		if _, err = kubernetes.GetNode("uyuni"); err != nil {
			err = kubernetes.ReplicasTo(kubernetes.ServerFilter, 1)
		}
	}()
	if inspectedValues["image_pg_version"] > inspectedValues["current_pg_version"] {
		log.Info().Msgf(L("Previous PostgreSQL is %s, new one is %s. Performing a DB version upgrade..."), inspectedValues["current_pg_version"], inspectedValues["image_pg_version"])

		if err := RunPgsqlVersionUpgrade(*image, *migrationImage, nodeName, inspectedValues["current_pg_version"], inspectedValues["image_pg_version"]); err != nil {
			return fmt.Errorf(L("cannot run PostgreSQL version upgrade script: %s"), err)
		}
	} else if inspectedValues["image_pg_version"] == inspectedValues["current_pg_version"] {
		log.Info().Msgf(L("Upgrading to %s without changing PostgreSQL version"), inspectedValues["uyuni_release"])
	} else {
		return fmt.Errorf(L("trying to downgrade PostgreSQL from %s to %s"), inspectedValues["current_pg_version"], inspectedValues["image_pg_version"])
	}

	schemaUpdateRequired := inspectedValues["current_pg_version"] != inspectedValues["image_pg_version"]
	if err := RunPgsqlFinalizeScript(serverImage, image.PullPolicy, nodeName, schemaUpdateRequired); err != nil {
		return fmt.Errorf(L("cannot run PostgreSQL version upgrade script: %s"), err)
	}

	if err := RunPostUpgradeScript(serverImage, image.PullPolicy, nodeName); err != nil {
		return fmt.Errorf(L("cannot run post upgrade script: %s"), err)
	}

	err = UyuniUpgrade(serverImage, image.PullPolicy, &helm, kubeconfig, fqdn, clusterInfos.Ingress)
	if err != nil {
		return fmt.Errorf(L("cannot upgrade to image %s: %s"), serverImage, err)
	}

	return kubernetes.WaitForDeployment(helm.Uyuni.Namespace, "uyuni", "uyuni")
}
