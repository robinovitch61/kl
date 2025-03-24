package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
	"github.com/robinovitch61/kl/internal/model"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"strings"
	"time"
)

// K8sClient is an interface for interacting with a Kubernetes cluster
type K8sClient interface {
	// AllClusterNamespaces returns all cluster namespaces
	AllClusterNamespaces() []k8s_model.ClusterNamespaces

	// GetContainerListener returns a listener that emits container deltas for a given cluster and namespace
	GetContainerListener(
		cluster,
		namespace string,
		matchers model.Matchers,
		selector labels.Selector,
		ignorePodOwnerTypes []string,
	) (ContainerListener, error)

	// CollectContainerDeltasForDuration collects container deltas from a listener for a given duration
	CollectContainerDeltasForDuration(listener ContainerListener, duration time.Duration) (container.ContainerDeltaSet, error)

	// GetContainerStatus returns the status of a container
	GetContainerStatus(container container.Container) (container.ContainerStatus, error)

	// GetLogStream returns a scanner that reads lines from a container's log stream
	GetLogStream(container container.Container, sinceTime time.Time) (*bufio.Scanner, context.CancelFunc, error)
}

type clientImpl struct {
	ctx                  context.Context
	clusterToClientset   map[string]*kubernetes.Clientset
	allClusterNamespaces []k8s_model.ClusterNamespaces
}

func (c clientImpl) AllClusterNamespaces() []k8s_model.ClusterNamespaces {
	return c.allClusterNamespaces
}

type ContainerListener struct {
	Cluster            string
	Namespace          string
	Stop               func()
	containerDeltaChan chan container.ContainerDelta
}

func (c clientImpl) GetContainerListener(
	cluster,
	namespace string,
	matchers model.Matchers,
	selector labels.Selector,
	ignorePodOwnerTypes []string,
) (ContainerListener, error) {
	deltaChan := make(chan container.ContainerDelta, 100)
	stopChan := make(chan struct{})

	// every 10 minutes, informer will resync, emitting new events for all discrepancies
	factory := informers.NewSharedInformerFactoryWithOptions(
		c.clusterToClientset[cluster],
		10*time.Minute,
		informers.WithNamespace(namespace),
	)

	podInformer := factory.Core().V1().Pods().Informer()

	_, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			deltas := getContainerDeltas(pod, cluster, false, matchers, selector, ignorePodOwnerTypes)
			for _, delta := range deltas {
				dev.Debug(fmt.Sprintf("listener add container %s, state %s", delta.Container.HumanReadable(), delta.Container.Status.State))
				deltaChan <- delta
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*corev1.Pod)
			if !ok {
				return
			}
			deltas := getContainerDeltas(pod, cluster, false, matchers, selector, ignorePodOwnerTypes)
			for _, delta := range deltas {
				dev.Debug(fmt.Sprintf("listener update container %s, state %s", delta.Container.HumanReadable(), delta.Container.Status.State))
				deltaChan <- delta
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			deltas := getContainerDeltas(pod, cluster, true, matchers, selector, ignorePodOwnerTypes)

			// sometimes the listener will receive a delete event for pods whose container statuses are not terminated
			// since we keep these around for a while, manually override the status to terminated
			for i := range deltas {
				deltas[i].Container.Status.State = container.ContainerTerminated
				deltas[i].Container.Status.StartedAt = time.Time{}
			}

			for _, delta := range deltas {
				dev.Debug(fmt.Sprintf("listener delete container %s, state %s", delta.Container.HumanReadable(), delta.Container.Status.State))
				deltaChan <- delta
			}
		},
	})
	if err != nil {
		return ContainerListener{}, fmt.Errorf("error adding event handler: %v", err)
	}

	go func() {
		podInformer.Run(stopChan)
	}()

	if !cache.WaitForCacheSync(stopChan, podInformer.HasSynced) {
		close(stopChan)
		return ContainerListener{}, fmt.Errorf("timed out waiting for caches to sync")
	}

	stop := func() {
		close(stopChan)
		close(deltaChan)
	}

	return ContainerListener{
		Cluster:            cluster,
		Namespace:          namespace,
		containerDeltaChan: deltaChan,
		Stop:               stop,
	}, nil
}

func (c clientImpl) CollectContainerDeltasForDuration(
	listener ContainerListener,
	duration time.Duration,
) (container.ContainerDeltaSet, error) {
	var deltas container.ContainerDeltaSet
	timeout := time.After(duration)

	for {
		select {
		case containerDelta, ok := <-listener.containerDeltaChan:
			if !ok {
				return container.ContainerDeltaSet{}, fmt.Errorf("add/update pod channel closed")
			}
			deltas.Add(containerDelta)

		case <-timeout:
			return deltas, nil
		}
	}
}

func (c clientImpl) GetContainerStatus(
	ct container.Container,
) (container.ContainerStatus, error) {
	clientset := c.clusterToClientset[ct.Cluster]
	if clientset == nil {
		return container.ContainerStatus{}, fmt.Errorf("clientset for cluster %s not found", ct.Cluster)
	}

	pod, err := clientset.CoreV1().Pods(ct.Namespace).Get(c.ctx, ct.Pod, metav1.GetOptions{})
	if err != nil {
		return container.ContainerStatus{}, fmt.Errorf("error getting pod %s in namespace %s: %v", ct.Pod, ct.Namespace, err)
	}
	return getStatus(pod.Status.ContainerStatuses, ct.Name)
}

func (c clientImpl) GetLogStream(
	container container.Container,
	sinceTime time.Time,
) (*bufio.Scanner, context.CancelFunc, error) {
	clientset := c.clusterToClientset[container.Cluster]
	if clientset == nil {
		return nil, nil, fmt.Errorf("clientset for cluster %s not found", container.Cluster)
	}

	logOptions := &v1.PodLogOptions{
		Container:  container.Name,
		Timestamps: true,
		Follow:     true,
		SinceTime:  &metav1.Time{Time: sinceTime},
	}
	logs := clientset.CoreV1().Pods(container.Namespace).GetLogs(container.Pod, logOptions)
	childCtx, cancel := context.WithCancel(c.ctx)
	logStream, err := logs.Stream(childCtx)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("error getting log scanner: %v", err)
	}

	// create a scanner that reads lines from the log stream
	scanner := bufio.NewScanner(logStream)
	maxLineLength := 1024 * 1024 * 1024
	scanner.Buffer(make([]byte, 64*1024), maxLineLength)
	scanner.Split(bufio.ScanLines)
	return scanner, cancel, nil
}

func getContainerDeltas(
	pod *corev1.Pod,
	cluster string,
	delete bool,
	matchers model.Matchers,
	selector labels.Selector,
	ignorePodOwnerTypes []string,
) []container.ContainerDelta {
	if pod == nil {
		return nil
	}
	now := time.Now()
	var deltas []container.ContainerDelta
	containers := getContainers(*pod, cluster, ignorePodOwnerTypes)
	for i := range containers {
		if matchers.IgnoreMatcher.MatchesContainer(containers[i]) {
			continue
		}
		matcherSelectsContainer := matchers.AutoSelectMatcher.MatchesContainer(containers[i])
		labelSelectorSelectsContainer := !selector.Empty() && selector.Matches(labels.Set(pod.Labels))
		delta := container.ContainerDelta{
			Time:       now,
			Container:  containers[i],
			ToDelete:   delete,
			ToActivate: matcherSelectsContainer || labelSelectorSelectsContainer,
		}
		deltas = append(deltas, delta)
	}
	return deltas
}

func getContainers(pod corev1.Pod, cluster string, ignorePodOwnerTypes []string) []container.Container {
	var containers []container.Container

	podOwnerName, ownerRefType := getPodOwnerNameAndOwnerRefType(pod)
	for _, ignored := range ignorePodOwnerTypes {
		if ignored != "" && ownerRefType == ignored {
			dev.Debug(fmt.Sprintf("ignoring pod %s with owner refs %+v", pod.Name, pod.OwnerReferences))
			return containers
		}
	}
	if podOwnerName == "" {
		dev.Debug(fmt.Sprintf("skipping pod %s with owner refs %+v", pod.Name, pod.OwnerReferences))
		return containers
	}

	metadata := k8s_model.PodOwnerMetadata{OwnerType: ownerRefType}

	for _, c := range pod.Spec.Containers {
		status, _ := getStatus(pod.Status.ContainerStatuses, c.Name)
		newContainer := container.Container{
			Cluster:          cluster,
			Namespace:        pod.Namespace,
			PodOwner:         podOwnerName,
			Pod:              pod.Name,
			Name:             c.Name,
			Status:           status,
			PodOwnerMetadata: metadata,
		}
		containers = append(containers, newContainer)
	}
	return containers
}

func getPodOwnerNameAndOwnerRefType(pod corev1.Pod) (string, string) {
	if len(pod.OwnerReferences) == 0 {
		return "unowned", "Unowned"
	}

	// ignore the fact that pods may have multiple owners for now
	podOwnerRef := pod.OwnerReferences[0]
	if strings.ToLower(podOwnerRef.Kind) == "replicaset" {
		// assume naming convention is <deployment-name>-<replica-set-hash>
		parts := strings.Split(podOwnerRef.Name, "-")
		if len(parts) > 1 {
			return strings.Join(parts[:len(parts)-1], "-"), "Deployment"
		}
	}

	// assume name is itself the pod owner name
	return podOwnerRef.Name, podOwnerRef.Kind
}

func getStatus(podContainerStatuses []v1.ContainerStatus, containerName string) (container.ContainerStatus, error) {
	for _, status := range podContainerStatuses {
		if status.Name == containerName {
			state, err := getState(status)
			if err != nil {
				return container.ContainerStatus{}, err
			}

			var startedAt time.Time
			var terminatedAt time.Time
			var waitingFor, terminatedFor string
			switch state {
			case container.ContainerRunning:
				if status.State.Running != nil {
					startedAt = status.State.Running.StartedAt.Time
				}
			case container.ContainerTerminated:
				if status.State.Terminated != nil {
					startedAt = status.State.Terminated.StartedAt.Time
					terminatedAt = status.State.Terminated.FinishedAt.Time
					terminatedFor = status.State.Terminated.Reason
				}
			case container.ContainerWaiting:
				if status.State.Waiting != nil {
					waitingFor = status.State.Waiting.Reason
				}
			default:
			}

			return container.ContainerStatus{
				State:         state,
				StartedAt:     startedAt,
				TerminatedAt:  terminatedAt,
				WaitingFor:    waitingFor,
				TerminatedFor: terminatedFor,
			}, nil
		}
	}
	return container.ContainerStatus{}, fmt.Errorf("container %s status not found", containerName)
}

func getState(status corev1.ContainerStatus) (container.ContainerState, error) {
	if status.State.Running != nil {
		return container.ContainerRunning, nil
	}
	if status.State.Terminated != nil {
		return container.ContainerTerminated, nil
	}
	if status.State.Waiting != nil {
		return container.ContainerWaiting, nil
	}
	return container.ContainerUnknown, fmt.Errorf("unknown container status %+v", status)
}
