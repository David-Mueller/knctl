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

package logs

import (
	"sync"

	"github.com/cppforlife/go-cli-ui/ui"
	corev1 "k8s.io/api/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PodLogOpts struct {
	Follow bool
	Lines  *int64
}

type PodLog struct {
	pod        corev1.Pod
	podsClient typedcorev1.PodInterface

	tag  string
	opts PodLogOpts
}

func NewPodLog(
	pod corev1.Pod,
	podsClient typedcorev1.PodInterface,
	tag string,
	opts PodLogOpts,
) PodLog {
	return PodLog{pod, podsClient, tag, opts}
}

// TailAll will tail all logs from all containers in a single Pod
func (l PodLog) TailAll(ui ui.UI, cancelCh chan struct{}) error {
	// Container will not emit any new logs since this is a terminal position
	podInTerminalState := l.pod.Status.Phase == corev1.PodSucceeded || l.pod.Status.Phase == corev1.PodFailed

	var conts []corev1.Container

	for _, cont := range l.pod.Spec.InitContainers {
		if !(podInTerminalState && l.isWaitingContainer(cont, l.pod.Status.InitContainerStatuses)) {
			conts = append(conts, cont)
		}
	}

	for _, cont := range l.pod.Spec.Containers {
		if !(podInTerminalState && l.isWaitingContainer(cont, l.pod.Status.ContainerStatuses)) {
			conts = append(conts, cont)
		}
	}

	var wg sync.WaitGroup

	for _, cont := range conts {
		cont := cont
		wg.Add(1)

		go func() {
			NewPodContainerLog(l.pod, cont.Name, l.podsClient, l.tag, l.opts).Tail(ui, cancelCh) // TODO err?
			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}

func (l PodLog) isWaitingContainer(cont corev1.Container, statuses []corev1.ContainerStatus) bool {
	for _, contStatus := range statuses {
		if cont.Name == contStatus.Name {
			return contStatus.State.Waiting != nil
		}
	}
	return false
}
