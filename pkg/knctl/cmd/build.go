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

package cmd

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlbuild "github.com/cppforlife/knctl/pkg/knctl/build"
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BuildOptions struct {
	ui            ui.UI
	depsFactory   DepsFactory
	cancelSignals CancelSignals

	BuildFlags       BuildFlags
	BuildCreateFlags BuildCreateFlags
}

func NewBuildOptions(ui ui.UI, depsFactory DepsFactory, cancelSignals CancelSignals) *BuildOptions {
	return &BuildOptions{ui: ui, depsFactory: depsFactory, cancelSignals: cancelSignals}
}

func NewBuildCmd(o *BuildOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build source code into image",
		Example: `
  # Build Git repository into an image in namespace 'ns1'
  knctl build -b build1 --git-url github.com/cppforlife/simple-app --git-revision master/head -i docker.io/cppforlife/simple-app -n ns1`, // TODO replace example
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.BuildFlags.Set(cmd)
	o.BuildCreateFlags.Set(cmd)
	return cmd
}

func (o *BuildOptions) Run() error {
	buildClient, err := o.depsFactory.BuildClient()
	if err != nil {
		return err
	}

	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	buildsClient := buildClient.BuildV1alpha1().Builds(o.BuildFlags.NamespaceFlags.Name)

	build := &buildv1alpha1.Build{
		ObjectMeta: o.BuildCreateFlags.GenerateNameFlags.Apply(metav1.ObjectMeta{
			Name:      o.BuildFlags.Name,
			Namespace: o.BuildFlags.NamespaceFlags.Name,
		}),
		Spec: ctlbuild.BuildSpec{}.Build(o.BuildCreateFlags.BuildSpecOpts),
	}

	build, err = buildsClient.Create(build)
	if err != nil {
		return err // TODO allow updating build?
	}

	cancelCh := make(chan struct{})
	o.cancelSignals.Watch(func() { close(cancelCh) })

	buildObj := ctlbuild.NewBuild(build, buildsClient, coreClient.CoreV1())

	err = buildObj.TailLogs(o.ui, cancelCh)
	if err != nil {
		return err
	}

	return buildObj.Error(cancelCh)
}
