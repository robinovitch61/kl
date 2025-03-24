package internal

import (
	"context"
	"flag"
	"github.com/robinovitch61/kl/internal/k8s/entity"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/command"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/k8s/client"
	"github.com/robinovitch61/kl/internal/message"
	"github.com/robinovitch61/kl/internal/page"
	"github.com/robinovitch61/kl/internal/style"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // register OIDC auth provider
	"k8s.io/klog/v2"
)

func initializedModel(m Model) (Model, tea.Cmd, error) {
	dev.Debug("initializing")
	defer dev.Debug("done initializing")
	dev.Debug("------------")

	// disable kubernetes client warning/logging
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "false")
	_ = flag.Set("stderrthreshold", "FATAL") // Set threshold to FATAL to suppress most kubernetes client logs
	_ = flag.Set("v", "0")                   // Set verbosity level to 0

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	c, err := client.NewK8sClient(
		ctx,
		m.config.KubeConfigPath,
		m.config.Contexts,
		m.config.Namespaces,
		m.config.AllNamespaces,
	)
	if err != nil {
		return m, nil, err
	}
	m.k8sClient = c

	m.components.entityTree = entity.NewEntityTree(m.k8sClient.AllClusterNamespaces())

	m.data.termStyleData = style.NewTermStyleData()

	m = initializePages(m)

	cmds := createInitialCommands(m)

	return m, tea.Batch(cmds...), nil
}

func initializePages(m Model) Model {
	if m.config.LogsView {
		m.state.focusedPageType = page.LogsPageType
	} else {
		m.state.focusedPageType = page.EntitiesPageType
	}
	m.state.rightPageType = page.LogsPageType
	m.state.sinceTime = m.config.SinceTime

	m.pages = make(map[page.Type]page.GenericPage)

	m.data.topBarHeight = lipgloss.Height(m.topBar())
	contentHeight := m.state.height - m.data.topBarHeight
	// keep all pages unfocused here since first page focus happens when first containers received
	m.pages[page.EntitiesPageType] = page.NewEntitiesPage(m.keyMap, m.state.width, contentHeight, m.components.entityTree, style.Styles{})
	m.pages[page.LogsPageType] = page.NewLogsPage(m.keyMap, m.state.width, contentHeight, m.config.Descending, style.Styles{})
	m.pages[page.SingleLogPageType] = page.NewSingleLogPage(m.keyMap, m.state.width, contentHeight, style.Styles{})

	if m.config.LogFilter.Value != "" {
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithLogFilter(m.config.LogFilter)
	}

	m.state.initialized = true
	return m
}

func createInitialCommands(m Model) []tea.Cmd {
	var cmds []tea.Cmd
	for _, clusterNamespaces := range m.k8sClient.AllClusterNamespaces() {
		for _, namespace := range clusterNamespaces.Namespaces {
			cmds = append(cmds, command.GetContainerListenerCmd(
				m.k8sClient,
				clusterNamespaces.Cluster,
				namespace,
				m.config.Matchers,
				m.config.Selector,
				m.config.IgnoreOwnerTypes,
			))
		}
	}

	updateSinceTimeTextCmd := tea.Tick(
		m.state.sinceTime.TimeToNextUpdate(),
		func(t time.Time) tea.Msg { return message.UpdateSinceTimeTextMsg{UUID: m.state.sinceTime.UUID} },
	)
	cmds = append(cmds, updateSinceTimeTextCmd)

	return cmds
}
