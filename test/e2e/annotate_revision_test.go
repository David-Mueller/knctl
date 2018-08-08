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
	"fmt"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"gopkg.in/yaml.v2"
)

func TestAnnotateRevision(t *testing.T) {
	logger := Logger{}
	knctl := Knctl{t, logger}

	const (
		serviceName         = "test-annotate-revisions-service-name"
		expectedContentRev1 = "TestRevisions_ContentRev1"
		expectedContentRev2 = "TestRevisions_ContentRev2"
		expectedContentRev3 = "TestRevisions_ContentRev3"
	)

	logger.Section("Delete previous service with the same name if exists", func() {
		knctl.RunWithErr([]string{"delete", "service", "-n", "default", "-s", serviceName})
	})

	defer func() {
		knctl.RunWithErr([]string{"delete", "service", "-n", "default", "-s", serviceName})
	}()

	logger.Section("Deploy 3 revisions", func() {
		knctl.Run([]string{
			"deploy",
			"-n", "default",
			"-s", serviceName,
			"-i", "gcr.io/knative-samples/helloworld-go",
			"-e", "TARGET=" + expectedContentRev1,
		})

		knctl.Run([]string{
			"deploy",
			"-n", "default",
			"-s", serviceName,
			"-i", "gcr.io/knative-samples/helloworld-go",
			"-e", "TARGET=" + expectedContentRev2,
		})

		knctl.Run([]string{
			"deploy",
			"-n", "default",
			"-s", serviceName,
			"-i", "gcr.io/knative-samples/helloworld-go",
			"-e", "TARGET=" + expectedContentRev3,
		})
	})

	logger.Section("Wait for 3 revisions", func() {
		var success bool

		for i := 0; i < 30; i++ {
			out := knctl.Run([]string{"list", "revisions", "-n", "default", "-s", serviceName, "--json"})
			resp := uitest.JSONUIFromBytes(t, []byte(out))

			if len(resp.Tables[0].Rows) == 3 {
				success = true
				break
			}
		}

		if !success {
			t.Fatalf("Expected to find 3 revisions")
		}
	})

	const (
		annotationKey           = "custom-key"
		annotationValue         = "custom-val"
		annotationCustomNameKey = "custom-name"
	)

	logger.Section("Checking that there are no annotations", func() {
		out := knctl.Run([]string{"list", "revisions", "-n", "default", "-s", serviceName, "--json"})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		if len(resp.Tables[0].Rows) != 3 {
			t.Fatalf("Expected to see 3 revisions in the list of revisions, but did not: '%s'", out)
		}

		for _, row := range resp.Tables[0].Rows {
			var anns map[string]interface{}

			err := yaml.Unmarshal([]byte(row["annotations"]), &anns)
			if err != nil {
				t.Fatalf("Expected YAML unmarshaling to succeed: '%s'", err)
			}

			if _, found := anns[annotationKey]; found {
				t.Fatalf("Did not expect to find annotation in '%#v'", anns)
			}
		}
	})

	logger.Section("Annotating revisions", func() {
		out := knctl.Run([]string{"list", "revisions", "-n", "default", "-s", serviceName, "--json"})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		for _, row := range resp.Tables[0].Rows {
			ann1 := fmt.Sprintf("%s=%s", annotationKey, annotationValue)
			ann2 := fmt.Sprintf("%s=%s", annotationCustomNameKey, row["name"])
			knctl.Run([]string{"annotate", "revision", "-n", "default", "-r", row["name"], "-a", ann1, "-a", ann2})
		}
	})

	logger.Section("Checking that there are annotations", func() {
		out := knctl.Run([]string{"list", "revisions", "-n", "default", "-s", serviceName, "--json"})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		for _, row := range resp.Tables[0].Rows {
			var anns map[string]interface{}

			err := yaml.Unmarshal([]byte(row["annotations"]), &anns)
			if err != nil {
				t.Fatalf("Expected YAML unmarshaling to succeed: '%s'", err)
			}

			if anns[annotationKey] != annotationValue {
				t.Fatalf("Expected revision to be annotated, but was not '%#v'", anns)
			}
			if anns[annotationCustomNameKey] != row["name"] {
				t.Fatalf("Expected revision to be annotated with a second annotation, but was not '%#v'", anns)
			}
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
