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

package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CurlOptions struct {
	ui          ui.UI
	depsFactory DepsFactory

	ServiceFlags ServiceFlags
}

func NewCurlOptions(ui ui.UI, depsFactory DepsFactory) *CurlOptions {
	return &CurlOptions{ui: ui, depsFactory: depsFactory}
}

func NewCurlCmd(o *CurlOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "curl",
		Short: "Curl service",
		Long: `Send a HTTP request to the first ingress address with the Host header set to the service's domain.

Requires 'curl' command installed on the system.`,
		Example: `
  # Curl service 'svc1' in namespace 'ns1'
  knctl curl -s svc1 -n ns1`,
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.ServiceFlags.Set(cmd)
	return cmd
}

func (o *CurlOptions) Run() error {
	serviceDomain, err := o.serviceDomain()
	if err != nil {
		return err
	}

	ingressAddress, err := o.ingressAddress()
	if err != nil {
		return err
	}

	cmdName := "curl"
	cmdArgs := []string{"-H", "Host: " + serviceDomain, "http://" + ingressAddress} // TODO Determine protocol for the entrypoint

	o.ui.PrintLinef("Running: %s '%s'", cmdName, strings.Join(cmdArgs, "' '"))

	out, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return fmt.Errorf("Running curl: %s", err)
	}

	o.ui.PrintBlock(out)

	return nil
}

func (o *CurlOptions) serviceDomain() (string, error) {
	servingClient, err := o.depsFactory.ServingClient()
	if err != nil {
		return "", err
	}

	service, err := servingClient.ServingV1alpha1().Services(o.ServiceFlags.NamespaceFlags.Name).Get(o.ServiceFlags.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if len(service.Status.Domain) == 0 {
		return "", fmt.Errorf("Expected service '%s' to have non-empty domain", o.ServiceFlags.Name)
	}

	return service.Status.Domain, nil
}

func (o *CurlOptions) ingressAddress() (string, error) {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return "", err
	}

	services, err := IngressServices{coreClient}.List()
	if err != nil {
		return "", err
	}

	for _, svc := range services.Items {
		addrs := []string{}
		ports := []int32{}

		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if len(ing.IP) > 0 {
				addrs = append(addrs, ing.IP)
			}
			if len(ing.Hostname) > 0 {
				addrs = append(addrs, ing.Hostname)
			}
		}

		for _, port := range svc.Spec.Ports {
			ports = append(ports, port.Port)
		}

		if len(addrs) > 0 && len(ports) > 0 {
			return fmt.Sprintf("%s:%d", addrs[0], ports[0]), nil
		}
	}

	return "", fmt.Errorf("Expected to find at least one ingress address")
}
