package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/robinovitch61/kl/internal/dev"
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

// Client is an interface for interacting with a Kubernetes cluster
type Client interface {
	// GetContainerListener returns a listener that emits container deltas for a given cluster and namespace
	GetContainerListener(
		cluster,
		namespace string,
		matchers model.Matchers,
		selector labels.Selector,
		ignorePodOwnerTypes []string,
	) (model.ContainerListener, error)

	// CollectContainerDeltasForDuration collects container deltas from a listener for a given duration
	CollectContainerDeltasForDuration(listener model.ContainerListener, duration time.Duration) (model.ContainerDeltaSet, error)

	// GetContainerStatus returns the status of a container
	GetContainerStatus(container model.Container) (model.ContainerStatus, error)

	// GetLogStream returns a scanner that reads lines from a container's log stream
	GetLogStream(container model.Container, sinceTime time.Time) (*bufio.Scanner, context.CancelFunc, error)
}

type clientImpl struct {
	ctx                context.Context
	clusterToClientset map[string]*kubernetes.Clientset
}

func NewClient(ctx context.Context, clusterToClientset map[string]*kubernetes.Clientset) Client {
	return clientImpl{
		ctx:                ctx,
		clusterToClientset: clusterToClientset,
	}
}

func (c clientImpl) GetContainerListener(
	cluster,
	namespace string,
	matchers model.Matchers,
	selector labels.Selector,
	ignorePodOwnerTypes []string,
) (model.ContainerListener, error) {
	deltaChan := make(chan model.ContainerDelta, 100)
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
				deltas[i].Container.Status.State = model.ContainerTerminated
				deltas[i].Container.Status.StartedAt = time.Time{}
			}

			for _, delta := range deltas {
				dev.Debug(fmt.Sprintf("listener delete container %s, state %s", delta.Container.HumanReadable(), delta.Container.Status.State))
				deltaChan <- delta
			}
		},
	})
	if err != nil {
		return model.ContainerListener{}, fmt.Errorf("error adding event handler: %v", err)
	}

	go func() {
		podInformer.Run(stopChan)
	}()

	if !cache.WaitForCacheSync(stopChan, podInformer.HasSynced) {
		close(stopChan)
		return model.ContainerListener{}, fmt.Errorf("timed out waiting for caches to sync")
	}

	stop := func() {
		close(stopChan)
		close(deltaChan)
	}

	return model.ContainerListener{
		Cluster:            cluster,
		Namespace:          namespace,
		ContainerDeltaChan: deltaChan,
		Stop:               stop,
	}, nil
}

func (c clientImpl) CollectContainerDeltasForDuration(
	listener model.ContainerListener,
	duration time.Duration,
) (model.ContainerDeltaSet, error) {
	var deltas model.ContainerDeltaSet
	timeout := time.After(duration)

	for {
		select {
		case containerDelta, ok := <-listener.ContainerDeltaChan:
			if !ok {
				return model.ContainerDeltaSet{}, fmt.Errorf("add/update pod channel closed")
			}
			deltas.Add(containerDelta)

		case <-timeout:
			return deltas, nil
		}
	}
}

func (c clientImpl) GetContainerStatus(
	container model.Container,
) (model.ContainerStatus, error) {
	clientset := c.clusterToClientset[container.Cluster]
	if clientset == nil {
		return model.ContainerStatus{}, fmt.Errorf("clientset for cluster %s not found", container.Cluster)
	}

	pod, err := clientset.CoreV1().Pods(container.Namespace).Get(c.ctx, container.Pod, metav1.GetOptions{})
	if err != nil {
		return model.ContainerStatus{}, fmt.Errorf("error getting pod %s in namespace %s: %v", container.Pod, container.Namespace, err)
	}
	return getStatus(pod.Status.ContainerStatuses, container.Name)
}

func (c clientImpl) GetLogStream(
	container model.Container,
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
) []model.ContainerDelta {
	if pod == nil {
		return nil
	}
	now := time.Now()
	var deltas []model.ContainerDelta
	containers := getContainers(*pod, cluster, ignorePodOwnerTypes)
	for i := range containers {
		if matchers.IgnoreMatcher.MatchesContainer(containers[i]) {
			continue
		}
		matcherSelectsContainer := matchers.AutoSelectMatcher.MatchesContainer(containers[i])
		labelSelectorSelectsContainer := !selector.Empty() && selector.Matches(labels.Set(pod.Labels))
		delta := model.ContainerDelta{
			Time:       now,
			Container:  containers[i],
			ToDelete:   delete,
			ToActivate: matcherSelectsContainer || labelSelectorSelectsContainer,
		}
		deltas = append(deltas, delta)
	}
	return deltas
}

func getContainers(pod corev1.Pod, cluster string, ignorePodOwnerTypes []string) []model.Container {
	var containers []model.Container

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

	metadata := model.PodOwnerMetadata{OwnerType: ownerRefType}

	for _, container := range pod.Spec.Containers {
		status, _ := getStatus(pod.Status.ContainerStatuses, container.Name)
		newContainer := model.Container{
			Cluster:          cluster,
			Namespace:        pod.Namespace,
			PodOwner:         podOwnerName,
			Pod:              pod.Name,
			Name:             container.Name,
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

func getStatus(podContainerStatuses []v1.ContainerStatus, containerName string) (model.ContainerStatus, error) {
	for _, status := range podContainerStatuses {
		if status.Name == containerName {
			state, err := getState(status)
			if err != nil {
				return model.ContainerStatus{}, err
			}

			var startedAt time.Time
			var terminatedAt time.Time
			var waitingFor, terminatedFor string
			switch state {
			case model.ContainerRunning:
				if status.State.Running != nil {
					startedAt = status.State.Running.StartedAt.Time
				}
			case model.ContainerTerminated:
				if status.State.Terminated != nil {
					startedAt = status.State.Terminated.StartedAt.Time
					terminatedAt = status.State.Terminated.FinishedAt.Time
					terminatedFor = status.State.Terminated.Reason
				}
			case model.ContainerWaiting:
				if status.State.Waiting != nil {
					waitingFor = status.State.Waiting.Reason
				}
			default:
			}

			return model.ContainerStatus{
				State:         state,
				StartedAt:     startedAt,
				TerminatedAt:  terminatedAt,
				WaitingFor:    waitingFor,
				TerminatedFor: terminatedFor,
			}, nil
		}
	}
	return model.ContainerStatus{}, fmt.Errorf("container %s status not found", containerName)
}

func getState(status corev1.ContainerStatus) (model.ContainerState, error) {
	if status.State.Running != nil {
		return model.ContainerRunning, nil
	}
	if status.State.Terminated != nil {
		return model.ContainerTerminated, nil
	}
	if status.State.Waiting != nil {
		return model.ContainerWaiting, nil
	}
	return model.ContainerUnknown, fmt.Errorf("unknown container status %+v", status)
}
