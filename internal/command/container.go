package command

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/k8s"
	"github.com/robinovitch61/kl/internal/model"
	"k8s.io/apimachinery/pkg/labels"
	"time"
)

type GetContainerListenerMsg struct {
	Listener model.ContainerListener
	Err      error
}

func GetContainerListenerCmd(
	client k8s.Client,
	cluster, namespace string,
	matchers model.Matchers,
	selector labels.Selector,
	ignorePodOwnerTypes []string,
) tea.Cmd {
	return func() tea.Msg {
		listener, err := client.GetContainerListener(cluster, namespace, matchers, selector, ignorePodOwnerTypes)
		if err != nil {
			return GetContainerListenerMsg{
				Err: fmt.Errorf("error subscribing to cluster %s, namespace %s: %v", cluster, namespace, err),
			}
		}
		return GetContainerListenerMsg{
			Listener: listener,
		}
	}
}

type GetContainerDeltasMsg struct {
	Listener model.ContainerListener
	DeltaSet model.ContainerDeltaSet
	Err      error
}

func GetNextContainerDeltasCmd(
	client k8s.Client,
	listener model.ContainerListener,
	duration time.Duration,
) tea.Cmd {
	return func() tea.Msg {
		for {
			deltaSet, err := client.CollectContainerDeltasForDuration(listener, duration)
			if err != nil {
				return GetContainerDeltasMsg{
					Listener: listener,
					DeltaSet: model.ContainerDeltaSet{},
					Err:      err,
				}
			}
			if deltaSet.Size() > 0 {
				return GetContainerDeltasMsg{
					Listener: listener,
					DeltaSet: deltaSet,
				}
			}
		}
	}
}
