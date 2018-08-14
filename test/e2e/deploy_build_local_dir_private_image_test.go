/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeployBuildLocalDirPrivateImage(t *testing.T) {
	logger := Logger{}
	env := BuildEnv(t)
	knctl := Knctl{t, env.Namespace, logger}
	kubectl := Kubectl{t, env.Namespace, logger}
	curl := Curl{t, knctl}

	const (
		serviceName              = "test-d-b-p-i-l-d-service-name"
		pushPullDockerSecretName = serviceName + "-docker-secret"
		pullDockerSecretName     = serviceName + "-p-docker-secret"
		buildServiceAccountName  = serviceName + "-service-account"
		expectedContent1         = "TestDeployBuild_ContentV1"
		expectedContent2         = "TestDeployBuild_ContentV2"
	)

	cleanUp := func() {
		knctl.RunWithOpts([]string{"delete", "service", "-s", serviceName}, RunOpts{AllowError: true})
		kubectl.RunWithOpts([]string{"delete", "secret", pushPullDockerSecretName}, RunOpts{AllowError: true})
		kubectl.RunWithOpts([]string{"delete", "secret", pullDockerSecretName}, RunOpts{AllowError: true})
		kubectl.RunWithOpts([]string{"delete", "serviceaccount", buildServiceAccountName}, RunOpts{AllowError: true})
	}

	logger.Section("Delete previous service with the same name if exists", cleanUp)
	defer cleanUp()

	logger.Section("Add service account with Docker push secret", func() {
		knctl.RunWithOpts([]string{
			"create",
			"basic-auth-secret",
			"-s", pushPullDockerSecretName,
			"--docker-hub",
			"-u", env.BuildDockerUsername,
			"-p", env.BuildDockerPassword,
		}, RunOpts{Redact: true})

		knctl.RunWithOpts([]string{
			"create",
			"basic-auth-secret",
			"-s", pullDockerSecretName,
			"--docker-hub",
			"-u", env.BuildDockerUsername,
			"-p", env.BuildDockerPassword,
			"--for-pulling",
		}, RunOpts{Redact: true})

		knctl.Run([]string{
			"create",
			"service-account",
			"-a", buildServiceAccountName,
			"-s", pushPullDockerSecretName,
			"-s", pullDockerSecretName,
		})
	})

	cwdPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Expected not to fail getting current directory: %s", err)
	}

	sourceDir := filepath.Join(cwdPath, "assets", "simple-app")

	logger.Section("Deploy service v1", func() {
		knctl.Run([]string{
			"deploy",
			"-s", serviceName,
			"-d", sourceDir,
			"-i", env.BuildPrivateImage,
			"--service-account", buildServiceAccountName,
			"-e", "SIMPLE_MSG=" + expectedContent1,
		})

		curl.WaitForContent(serviceName, expectedContent1)
	})

	logger.Section("Deploy service v2 with a Git change (new env variable)", func() {
		knctl.Run([]string{
			"deploy",
			"-s", serviceName,
			"-d", sourceDir,
			"-i", env.BuildPrivateImage,
			"--service-account", buildServiceAccountName,
			"-e", "SIMPLE_MSG=" + expectedContent2,
		})

		curl.WaitForContent(serviceName, expectedContent2)
	})

	logger.Section("Deleting service", func() {
		knctl.Run([]string{"delete", "service", "-s", serviceName})

		out := knctl.Run([]string{"list", "services", "--json"})
		if strings.Contains(out, serviceName) {
			t.Fatalf("Expected to not see sample service in the list of services, but was: %s", out)
		}
	})
}
