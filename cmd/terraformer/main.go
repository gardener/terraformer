// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/terraformer/cmd/terraformer/app"
	"github.com/gardener/terraformer/pkg/terraformer"
	"github.com/gardener/terraformer/pkg/utils"
)

func checkNetworkConnectivity(url string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("unable to reach the server: %v", err)
	}
	defer resp.Body.Close()

	// Just return nil if the server is reachable, regardless of the status code
	return nil
}

func main() {
	// Check if the KUBERNETES_SERVICE_HOST environment variable is set
	kubernetesServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	if kubernetesServiceHost == "" {
		panic("KUBERNETES_SERVICE_HOST is not set. cannot start terraformer.")
	}

	// Retry network connectivity check every 5 seconds until successful
	url := fmt.Sprintf("https://%s", kubernetesServiceHost)
	for {
		if err := checkNetworkConnectivity(url); err != nil {
			fmt.Printf("network connectivity check failed: %v. Retrying in 5 seconds...\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	if err := exec.Command("which", terraformer.TerraformBinary).Run(); err != nil {
		panic("terraform is not installed or not executable. cannot start terraformer.")
	}

	if err := app.NewTerraformerCommand().Execute(); err != nil {
		if log := runtimelog.Log; log.Enabled() {
			log.Error(err, "error running terraformer")
		} else {
			fmt.Printf("error running terraformer: %v", err)
		}

		// set exit code from terraform
		withExitCode := &utils.WithExitCode{}
		if errors.As(err, withExitCode) {
			if exitCode := withExitCode.ExitCode(); exitCode > 0 {
				os.Exit(exitCode)
			}
		}
		os.Exit(1)
	}
}
