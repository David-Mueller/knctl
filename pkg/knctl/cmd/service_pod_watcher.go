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
	"sync"

	"github.com/cppforlife/go-cli-ui/ui"
	ctlservice "github.com/cppforlife/knctl/pkg/knctl/service"
	"github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type ServicePodWatcher struct {
	serviceNamespace string
	serviceName      string

	servingClient servingclientset.Interface
	coreClient    kubernetes.Interface

	ui ui.UI // TODO remove
}

func NewServicePodWatcher(
	serviceNamespace string,
	serviceName string,
	servingClient servingclientset.Interface,
	coreClient kubernetes.Interface,
	ui ui.UI,
) ServicePodWatcher {
	return ServicePodWatcher{serviceNamespace, serviceName, servingClient, coreClient, ui}
}

func (w ServicePodWatcher) Watch(podsToWatchCh chan corev1.Pod, cancelCh chan struct{}) error {
	nonUniqueRevisionsToWatchCh := make(chan v1alpha1.Revision)

	// Watch revisions in this service
	go func() {
		revisionWatcher := ctlservice.NewRevisionWatcher(
			w.servingClient.ServingV1alpha1().Revisions(w.serviceNamespace),
			metav1.ListOptions{
				LabelSelector: labels.Set(map[string]string{
					serving.ConfigurationLabelKey: w.serviceName,
				}).String(),
			},
		)

		err := revisionWatcher.Watch(nonUniqueRevisionsToWatchCh, cancelCh)
		if err != nil {
			w.ui.BeginLinef("Revision watching error: %s\n", err)
		}

		close(nonUniqueRevisionsToWatchCh)
	}()

	// Watch pods in each revision
	var wg sync.WaitGroup
	watchedRevs := map[string]struct{}{}

	for revision := range nonUniqueRevisionsToWatchCh {
		revision := revision

		revUID := string(revision.UID)
		if _, found := watchedRevs[revUID]; found {
			continue
		}

		watchedRevs[revUID] = struct{}{}
		wg.Add(1)

		go func() {
			err := NewRevisionPodWatcher(&revision, w.servingClient, w.coreClient, w.ui).Watch(podsToWatchCh, cancelCh)
			if err != nil {
				w.ui.BeginLinef("Pod watching error: %s\n", err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}
