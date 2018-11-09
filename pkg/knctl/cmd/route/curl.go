/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Open 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package route

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/cppforlife/knctl/pkg/knctl/cmd/core"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CurlOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	RouteFlags RouteFlags
	CurlFlags  CurlFlags
	Verbose    bool
}

func NewCurlOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *CurlOptions {
	return &CurlOptions{ui: ui, depsFactory: depsFactory}
}

func NewCurlCmd(o *CurlOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "curl",
		Short: "Curl route",
		Long: `Send a HTTP request to the first ingress address with the Host header set to the service's domain.

Requires 'curl' command installed on the system.`,
		Example: `
  # Curl route 'rt1' in namespace 'ns1'
  knctl route curl --route rt1 -n ns1`,
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.RouteFlags.Set(cmd, flagsFactory)
	o.CurlFlags.Set(cmd, flagsFactory)
	cmd.Flags().BoolVarP(&o.Verbose, "verbose", "v", false, "Makes curl verbose during the operation")
	return cmd
}

func (o *CurlOptions) Run() error {
	domain, url, err := o.addr()
	if err != nil {
		return err
	}

	cmdName := "curl"
	cmdArgs := []string{}

	if o.Verbose {
		cmdArgs = append(cmdArgs, "-vvv")
	}

	cmdArgs = append(cmdArgs, []string{"-sS", "-H", "Host: " + domain, url}...)

	o.ui.PrintLinef("Running: %s '%s'", cmdName, strings.Join(cmdArgs, "' '"))

	out, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running curl: %s", err)
	}

	o.ui.PrintBlock(out)

	return nil
}

func (o *CurlOptions) addr() (string, string, error) {
	servingClient, err := o.depsFactory.ServingClient()
	if err != nil {
		return "", "", err
	}

	route, err := servingClient.ServingV1alpha1().Routes(o.RouteFlags.NamespaceFlags.Name).Get(o.RouteFlags.Name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return "", "", err
	}

	routeAddr := RouteAddress{route, coreClient}

	domain, err := routeAddr.Domain()
	if err != nil {
		return "", "", err
	}

	url, err := routeAddr.URL(o.CurlFlags.Port, false)
	if err != nil {
		return "", "", err
	}

	return domain, url, nil
}
