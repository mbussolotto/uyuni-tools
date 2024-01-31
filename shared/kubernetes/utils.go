// SPDX-FileCopyrightText: 2023 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

const ServerFilter = "-lapp=uyuni"
const ProxyFilter = "-lapp=uyuni-proxy"

// waitForDeployment waits at most 60s for a kubernetes deployment to have at least one replica.
// See [isDeploymentReady] for more details.
func WaitForDeployment(namespace string, name string, appName string) {
	// Find the name of a replica pod
	// Using the app label is a shortcut, not the 100% acurate way to get from deployment to pod
	podName := ""
	jsonpath := fmt.Sprintf("jsonpath={.items[?(@.metadata.labels.app==\"%s\")].metadata.name}", appName)
	cmdArgs := []string{"get", "pod", "-o", jsonpath}
	cmdArgs = addNamespace(cmdArgs, namespace)

	for i := 0; i < 60; i++ {
		out, err := utils.RunCmdOutput(zerolog.DebugLevel, "kubectl", cmdArgs...)
		if err == nil {
			podName = string(out)
			break
		}
	}

	// We need to wait for the image to be pulled as this can add quite some time
	// Setting a timeout on this is very hard since it hightly depends on network speed and image size
	// List the Pulled events from the pod as we may not see the Pulling if the image was already downloaded
	WaitForPulledImage(namespace, podName)

	log.Info().Msgf("Waiting for %s deployment to be ready in %s namespace\n", name, namespace)
	// Wait for a replica to be ready
	for i := 0; i < 60; i++ {
		// TODO Look for pod failures
		if IsDeploymentReady(namespace, name) {
			return
		}
		time.Sleep(1 * time.Second)
	}
	log.Fatal().Msgf("Failed to find a ready replica for deployment %s in namespace %s after 60s", name, namespace)
}

func WaitForPulledImage(namespace string, podName string) {
	log.Info().Msgf("Waiting for image of %s pod in %s namespace to be pulled", podName, namespace)
	pulledArgs := []string{"get", "event",
		"-o", "jsonpath={.items[?(@.reason==\"Pulled\")].message}",
		"--field-selector", "involvedObject.name=" + podName}
	pulledArgs = addNamespace(pulledArgs, namespace)

	failedArgs := []string{"get", "event",
		"-o", "jsonpath={range .items[?(@.reason==\"Failed\")]}{.message}{\"\\n\"}{end}",
		"--field-selector", "involvedObject.name=" + podName}
	failedArgs = addNamespace(failedArgs, namespace)
	for {
		// Look for events indicating an image pull issue
		out, err := utils.RunCmdOutput(zerolog.TraceLevel, "kubectl", failedArgs...)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to get failed events for pod %s", podName)
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Failed to pull image") {
				log.Fatal().Err(err).Msg("Failed to pull image")
			}
		}

		// Has the image pull finished?
		out, err = utils.RunCmdOutput(zerolog.TraceLevel, "kubectl", pulledArgs...)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to get events for pod %s", podName)
		}
		if len(out) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

// IsDeploymentReady returns true if a kubernetes deployment has at least one ready replica.
// The name can also be a filter parameter like -lapp=uyuni.
// An empty namespace means searching through all the namespaces.
func IsDeploymentReady(namespace string, name string) bool {
	jsonpath := fmt.Sprintf("jsonpath={.items[?(@.metadata.name==\"%s\")].status.readyReplicas}", name)
	args := []string{"get", "-o", jsonpath, "deploy"}
	args = addNamespace(args, namespace)

	out, err := utils.RunCmdOutput(zerolog.TraceLevel, "kubectl", args...)
	// kubectl errors out if the deployment or namespace doesn't exist
	if err == nil {
		if replicas, _ := strconv.Atoi(string(out)); replicas > 0 {
			return true
		}
	}
	return false
}

func addNamespace(args []string, namespace string) []string {
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "-A")
	}
	return args
}

func GetPullPolicy(name string) string {
	policies := map[string]string{
		"always":       "Always",
		"never":        "Never",
		"ifnotpresent": "IfNotPresent",
	}
	policy := policies[strings.ToLower(name)]
	if policy == "" {
		log.Fatal().Msgf("%s is not a valid image pull policy value", name)
	}
	return policy
}

func SetReplicas(replica uint) bool {
	args := []string{"scale", "deploy", "uyuni", "--replicas"}

	args = append(args, fmt.Sprint(replica))

	out, err := utils.RunCmdOutput(zerolog.TraceLevel, "kubectl", args...)
	if err != nil {
		if replicas, _ := strconv.Atoi(string(out)); replicas > 0 {
			return true
		}
	}
	return false
}

func RunPod(podname string, image string, pullPolicy string, command string, args ...string) {
	arguments := []string{"run", podname, "--image", image, "--image-pull-policy", pullPolicy}

	arguments = append(arguments, args...)
	arguments = append(arguments, "--command", "--", command)
	err := utils.RunCmdStdMapping("kubectl", arguments...)

	if err != nil {
		log.Fatal().Err(err).Msgf("Cannot run %s using image %s", command, image)
	}
}

func DeletePod(podname string) string {
	arguments := []string{"delete", "pod", podname}

	out, err := utils.RunCmdOutput(zerolog.DebugLevel, "kubectl", arguments...)

	if err != nil {
		log.Fatal().Err(err).Msgf("Cannot %s", arguments)
	}
	return string(out)
}

func WaitForPod(podname string, status string) {
	waitSeconds := 60

	log.Debug().Msgf("Checking status for %s pod. Waiting %s seconds until status is %s", podname, strconv.Itoa(waitSeconds), status)

	cmdArgs := []string{"get", "pod", podname, "--output=custom-columns=STATUS:.status.phase", "--no-headers"}

	for i := 0; i < waitSeconds; i++ {
		out, err := utils.RunCmdOutput(zerolog.DebugLevel, "kubectl", cmdArgs...)
		log.Debug().Msgf("%s status is %s", podname, out)

		if strings.ToUpper(string(out)) == strings.ToUpper(status+"\n") {
			log.Debug().Msgf("%s pod status is %s", podname, status)
			break
		}
		if err != nil {
			log.Fatal().Err(err).Msgf("Pod %s does not %s in %s seconds", podname, out, strconv.Itoa(waitSeconds))
		}
		time.Sleep(1 * time.Second)
	}
}

func GetNode(appName string) string {
	nodeName := ""
	cmdArgs := []string{"get", "pod", "-lapp=" + appName, "-o", "jsonpath={.items[*].spec.nodeName}"}
	for i := 0; i < 60; i++ {
		out, err := utils.RunCmdOutput(zerolog.DebugLevel, "kubectl", cmdArgs...)
		if err == nil {
			nodeName = string(out)
			break
		}
	}
	if len(nodeName) > 0 {
		log.Debug().Msgf("Node name for app %s is: %s", appName, nodeName)
	} else {
		log.Warn().Msgf("Cannot find Node name for app %s", appName)
	}
	return nodeName
}
