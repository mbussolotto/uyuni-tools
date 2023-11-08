package podman

import (
	"strings"

	"github.com/rs/zerolog/log"
	podman_utils "github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
	"github.com/uyuni-project/uyuni-tools/uyuniproxy/shared/templates"
)

const commonArgs = "--rm --cap-add NET_RAW --tmpfs /run -v cgroup:/sys/fs/cgroup:rw"

func GetCommonParams() []string {
	return strings.Split(commonArgs, " ")
}

const ProxyPodPath = "/etc/systemd/system/uyuni-proxy.service"
const ProxyHttpdPath = "/etc/systemd/system/uyuni-proxy-httpd.service"
const ProxySaltBrokerPath = "/etc/systemd/system/uyuni-proxy-salt-broker.service,"
const ProxySquidPath = "/etc/systemd/system/uyuni-proxy-squid.service"
const ProxyTftpdPath = "/etc/systemd/system/uyuni-proxy-tftpd.service"

func GenerateSystemdService(image string, podmanArgs []string) {

	podman_utils.SetupNetwork()

	log.Info().Msg("Enabling system service")
	dataProxyPod := templates.PodmanProxyPodTemplateData{
		NamePrefix: "uyuni",
		Ports:      utils.PROXY_PORTS,
	}
	if err := utils.WriteTemplateToFile(dataProxyPod, ProxyPodPath, 0555, false); err != nil {
		log.Fatal().Err(err).Msgf("Failed to generate %s", ProxyPodPath)
	}

	dataHttpd := templates.PodmanProxyHttpdTemplateData{
		NamePrefix: "uyuni",
		Volumes:    utils.PROXY_HTTPD_VOLUMES,
	}
	if err := utils.WriteTemplateToFile(dataHttpd, ProxyHttpdPath, 0555, false); err != nil {
		log.Fatal().Err(err).Msgf("Failed to generate %s", ProxyHttpdPath)
	}

	dataProxySaltBroker := templates.PodmanProxySaltBrokerTemplateData{
		NamePrefix: "uyuni",
	}
	if err := utils.WriteTemplateToFile(dataProxySaltBroker, ProxySaltBrokerPath, 0555, false); err != nil {
		log.Fatal().Err(err).Msgf("Failed to generate %s", ProxySaltBrokerPath)
	}

	dataProxySquid := templates.PodmanProxySquidTemplateData{
		NamePrefix: "uyuni",
		Volumes:    utils.PROXY_SQUID_VOLUMES,
	}
	if err := utils.WriteTemplateToFile(dataProxySquid, ProxySquidPath, 0555, false); err != nil {
		log.Fatal().Err(err).Msgf("Failed to generate %s", ProxySquidPath)
	}

	dataProxyTftpd := templates.PodmanProxyTFTPDTemplateData{
		NamePrefix: "uyuni",
		Volumes:    utils.PROXY_TFTPD_VOLUMES,
	}
	if err := utils.WriteTemplateToFile(dataProxyTftpd, ProxyTftpdPath, 0555, false); err != nil {
		log.Fatal().Err(err).Msgf("Failed to generate %s", ProxyTftpdPath)
	}

	utils.RunCmd("systemctl", "daemon-reload")
}
