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
	"bufio"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PodContainerLog struct {
	pod        corev1.Pod
	container  string
	podsClient typedcorev1.PodInterface

	tag  string
	opts PodLogOpts
}

func NewPodContainerLog(
	pod corev1.Pod,
	container string,
	podsClient typedcorev1.PodInterface,
	tag string,
	opts PodLogOpts,
) PodContainerLog {
	return PodContainerLog{
		pod:        pod,
		container:  container,
		podsClient: podsClient,

		tag:  tag,
		opts: opts,
	}
}

func (l PodContainerLog) Tail(ui ui.UI, cancelCh chan struct{}) error {
	var streamCanceled atomic.Value

	stream, err := l.obtainStream(cancelCh)
	if err != nil {
		return err
	}

	if stream == nil {
		return nil
	}

	defer stream.Close()

	go func() {
		<-cancelCh
		streamCanceled.Store(true)
		stream.Close()
	}()

	reader := bufio.NewReader(stream)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			typedCanceled, ok := streamCanceled.Load().(bool)
			if ok && typedCanceled {
				return nil // ignore error if stream was canceled
			}
			return err
		}

		ui.PrintBlock([]byte(fmt.Sprintf("%s | %s", l.tag, line)))
	}
}

func (l PodContainerLog) obtainStream(cancelCh chan struct{}) (io.ReadCloser, error) {
	for {
		// TODO infinite retry

		// It appears that GetLogs will successfully return log stream
		// almost immediately after pod has been created; however,
		// returned log stream will not carry any data, even after containers have started.
		// Wait for pod object to have container status fields initialized
		// since that appears to make GetLogs call return actual log stream.
		if l.readyToGetLogs() {
			logs := l.podsClient.GetLogs(l.pod.Name, &corev1.PodLogOptions{
				Follow:    l.opts.Follow,
				TailLines: l.opts.Lines,
				Container: l.container,
				// TODO other options
			})

			stream, err := logs.Stream()
			if err == nil {
				return stream, nil
			}
		}

		select {
		case <-cancelCh:
			return nil, nil
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (l PodContainerLog) readyToGetLogs() bool {
	pod, err := l.podsClient.Get(l.pod.Name, metav1.GetOptions{})
	if err != nil {
		return false
	}

	return len(pod.Status.ContainerStatuses) > 0 || len(pod.Status.InitContainerStatuses) > 0
}
