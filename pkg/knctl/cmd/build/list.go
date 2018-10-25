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

package build

import (
	"fmt"
	cmdcore "github.com/cppforlife/knctl/pkg/knctl/cmd/core"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	NamespaceFlags cmdcore.NamespaceFlags
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: cmdcore.ListAliases,
		Short:   "List builds",
		Long:    "List all builds in a namespace",
		Example: `
  # List all builds in namespace 'ns1'
  knctl build list -n ns1`,
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *ListOptions) Run() error {
	buildClient, err := o.depsFactory.BuildClient()
	if err != nil {
		return err
	}

	builds, err := buildClient.BuildV1alpha1().Builds(o.NamespaceFlags.Name).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	table := uitable.Table{
		Title: fmt.Sprintf("Builds in namespace '%s'", o.NamespaceFlags.Name),

		Content: "builds",

		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Succeeded"),
			uitable.NewHeader("Age"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 2, Asc: false}, // Show latest first
		},
	}

	for _, build := range builds.Items {
		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(build.Name),
			NewBuildSucceededValue(build),
			cmdcore.NewValueAge(build.CreationTimestamp.Time),
		})
	}

	o.ui.PrintTable(table)

	return nil
}

func NewBuildSucceededValue(build v1alpha1.Build) cmdcore.ValueUnknownBool {
	cond := build.Status.GetCondition(v1alpha1.BuildSucceeded)
	if cond != nil {
		switch cond.Status {
		case corev1.ConditionTrue:
			result := true
			return cmdcore.NewValueUnknownBool(&result)
		case corev1.ConditionFalse:
			result := false
			return cmdcore.NewValueUnknownBool(&result)
		}
	}

	return cmdcore.NewValueUnknownBool(nil)
}
