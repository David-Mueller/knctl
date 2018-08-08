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
	"strings"
	"testing"
	"time"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestLogsFollow(t *testing.T) {
	logger := Logger{}
	knctl := Knctl{t, logger}
	curl := Curl{t, knctl}

	const (
		serviceName         = "test-logs-follow-service-name"
		expectedContentRev1 = "TestRevisions_ContentRev1"
		expectedContentRev2 = "TestRevisions_ContentRev2"
		expectedContentRev3 = "TestRevisions_ContentRev3"
	)

	logger.Section("Sleeping...", func() {
		// TODO otherwise 'no upstream healty' error happens
		// somehow caused by previous deploy in other tests
		time.Sleep(20 * time.Second)
	})

	logger.Section("Delete previous service with the same name if exists", func() {
		knctl.RunWithErr([]string{"delete", "service", "-n", "default", "-s", serviceName})
	})

	defer func() {
		knctl.RunWithErr([]string{"delete", "service", "-n", "default", "-s", serviceName})
	}()

	logger.Section("Deploy revision 1", func() {
		knctl.Run([]string{
			"deploy",
			"-n", "default",
			"-s", serviceName,
			"-i", "gcr.io/knative-samples/helloworld-go",
			"-e", "TARGET=" + expectedContentRev1,
		})
	})

	logger.Section("Checking if service is reachable and presents content", func() {
		curl.WaitForContent(serviceName, expectedContentRev1)
	})

	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})
	var collectedLogs string

	// Start tailing logs in the backgroud
	go func() {
		collectedLogs, _ = knctl.RunWithCancel([]string{"logs", "-n", "default", "-s", serviceName, "-f"}, cancelCh)
		doneCh <- struct{}{}
	}()

	logger.Section("Deploy revision 2 and check its logs", func() {
		knctl.Run([]string{
			"deploy",
			"-n", "default",
			"-s", serviceName,
			"-i", "gcr.io/knative-samples/helloworld-go",
			"-e", "TARGET=" + expectedContentRev2,
		})

		curl.WaitForContent(serviceName, expectedContentRev2)
	})

	logger.Section("Deploy revision 3 and check its logs", func() {
		knctl.Run([]string{
			"deploy",
			"-n", "default",
			"-s", serviceName,
			"-i", "gcr.io/knative-samples/helloworld-go",
			"-e", "TARGET=" + expectedContentRev3,
		})

		curl.WaitForContent(serviceName, expectedContentRev3)
	})

	cancelCh <- struct{}{}
	<-doneCh

	logger.Section("Check logs of service to make sure it includes logs from 3 revisions", func() {
		collectedLogsLines := strings.Split(collectedLogs, "\n")

		expectedLogLines := []string{
			"Hello world sample started.",
			"Hello world received a request.",
		}

		out := knctl.Run([]string{"list", "revisions", "-n", "default", "-s", serviceName, "--json"})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		if len(resp.Tables[0].Rows) != 3 {
			t.Fatalf("Expected to see one revision in the list of revisions, but did not: '%s'", out)
		}

		var matchedLines int

		for _, row := range resp.Tables[0].Rows {
			for _, expectedLogLine := range expectedLogLines {
				var found bool
				for _, line := range collectedLogsLines {
					if strings.HasPrefix(line, row["name"]+" >") && strings.HasSuffix(line, expectedLogLine) {
						found = true
						matchedLines++
						break
					}
				}
				if !found {
					t.Fatalf("Expected to find log line '%s' in service logs: '%s'", expectedLogLine, collectedLogs)
				}
			}
		}

		if matchedLines == 0 {
			t.Fatalf("Expected to have matched several lines")
		}
	})

	logger.Section("Deleting service", func() {
		knctl.Run([]string{"delete", "service", "-n", "default", "-s", serviceName})

		out := knctl.Run([]string{"list", "services", "-n", "default", "--json"})
		if strings.Contains(out, serviceName) {
			t.Fatalf("Expected to not see sample service in the list of services, but was: %s", out)
		}
	})
}
