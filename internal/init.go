package internal

import (
	"context"
	"flag"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/command"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/k8s/client"
	"github.com/robinovitch61/kl/internal/message"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/page"
	"github.com/robinovitch61/kl/internal/style"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // register OIDC auth provider
	"k8s.io/klog/v2"
	"time"
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

	// Currently intentionally unsupported in config, can revisit in the future but will make config more complicated:
	// - specifying a specific set of namespaces per context/cluster
	//    - could potentially edit `--contexts` to have form of `context1[ns1,ns2],context2[ns3,ns4]`
	// - specifying multiple contexts that point to the same cluster
	//    - I can't imagine a scenario where this is desired other than wanting multiple namespaces per cluster

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	c, err := client.NewClient(
		ctx,
		m.config.KubeConfigPath,
		m.config.Contexts,
		m.config.Namespaces,
		m.config.AllNamespaces,
	)
	if err != nil {
		return m, nil, err
	}
	m.client = c

	m.entityTree = model.NewEntityTree(m.client.AllClusterNamespaces())

	m.termStyleData = style.NewTermStyleData()

	m = initializePages(m)

	cmds := createInitialCommands(m)

	return m, tea.Batch(cmds...), nil
}

func initializePages(m Model) Model {
	if m.config.LogsView {
		m.focusedPageType = page.LogsPageType
	} else {
		m.focusedPageType = page.EntitiesPageType
	}
	m.rightPageType = page.LogsPageType
	m.sinceTime = m.config.SinceTime

	m.pages = make(map[page.Type]page.GenericPage)

	m.topBarHeight = lipgloss.Height(m.topBar())
	contentHeight := m.height - m.topBarHeight
	// keep all pages unfocused here since first page focus happens when first containers received
	m.pages[page.EntitiesPageType] = page.NewEntitiesPage(m.keyMap, m.width, contentHeight, m.entityTree, style.Styles{})
	m.pages[page.LogsPageType] = page.NewLogsPage(m.keyMap, m.width, contentHeight, m.config.Descending, style.Styles{})
	m.pages[page.SingleLogPageType] = page.NewSingleLogPage(m.keyMap, m.width, contentHeight, style.Styles{})

	if m.config.LogFilter.Value != "" {
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithLogFilter(m.config.LogFilter)
	}

	m.initialized = true
	return m
}

func createInitialCommands(m Model) []tea.Cmd {
	var cmds []tea.Cmd
	for _, clusterNamespaces := range m.client.AllClusterNamespaces() {
		for _, namespace := range clusterNamespaces.Namespaces {
			cmds = append(cmds, command.GetContainerListenerCmd(
				m.client,
				clusterNamespaces.Cluster,
				namespace,
				m.config.Matchers,
				m.config.Selector,
				m.config.IgnoreOwnerTypes,
			))
		}
	}

	updateSinceTimeTextCmd := tea.Tick(
		m.sinceTime.TimeToNextUpdate(),
		func(t time.Time) tea.Msg { return message.UpdateSinceTimeTextMsg{UUID: m.sinceTime.UUID} },
	)
	cmds = append(cmds, updateSinceTimeTextCmd)

	return cmds
}
