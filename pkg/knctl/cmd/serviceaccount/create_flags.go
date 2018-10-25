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

package serviceaccount

import (
	cmdcore "github.com/cppforlife/knctl/pkg/knctl/cmd/core"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	GenerateNameFlags cmdcore.GenerateNameFlags

	Secrets          []string
	ImagePullSecrets []string
}

func (s *CreateFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.GenerateNameFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringSliceVarP(&s.Secrets, "secret", "s", nil, "Set secret (format: secret-name) (can be specified multiple times)")
	cmd.Flags().StringSliceVarP(&s.ImagePullSecrets, "image-pull-secret", "p", nil, "Set image pull secret (format: secret-name) (can be specified multiple times)")
}
