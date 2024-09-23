package internal

// NOTE: Searching for `// #` will walk you through the main flow of the application

import (
	"context"
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
	"github.com/robinovitch61/kl/internal/command"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/fileio"
	"github.com/robinovitch61/kl/internal/k8s"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/message"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/page"
	"github.com/robinovitch61/kl/internal/prompt"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/toast"
	"github.com/robinovitch61/kl/internal/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

type Model struct {
	config               Config
	keyMap               keymap.KeyMap
	allClusterNamespaces []model.ClusterNamespaces
	width, height        int
	initialized          bool
	seenFirstContainer   bool
	toast                toast.Model
	prompt               prompt.Model
	whenPromptConfirm    func() (Model, tea.Cmd)
	err                  error
	entityTree           model.EntityTree
	containerToShortName func(model.Container) (string, error)
	pageLogBuffer        []model.PageLog
	client               k8s.Client
	cancel               context.CancelFunc
	pages                map[page.Type]page.GenericPage
	containerListeners   []model.ContainerListener
	currentPageType      page.Type
	sinceTime            model.SinceTime
	pendingSinceTime     *model.SinceTime
	pauseState           bool
	helpText             string
	topBarHeight         int // assumed constant
}

func InitialModel(c Config) Model {
	return Model{
		config: c,
		keyMap: keymap.DefaultKeyMap,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(constants.BatchUpdateLogsInterval, func(t time.Time) tea.Msg { return message.BatchUpdateLogsMsg{} }),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dev.DebugMsg("App", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case message.CleanupCompleteMsg:
		return m, tea.Quit

	// #5: The user presses a key. The key could be global, e.g. exit, or handled by the current page. Key presses update
	// models and can trigger commands to be run. Commands are used to perform background work, e.g. starting a log scanner.
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case message.ErrMsg:
		m.err = msg.Err

	// WindowSizeMsg arrives once on startup, then again every time the window is resized
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if !m.initialized {
			m, cmd = m.initialize()
			cmds = append(cmds, cmd)
		}
		m, cmd = m.handleWindowSizeMsg(msg.Width, msg.Height)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case message.AttemptUpdateSinceTimeMsg:
		m, cmd = m.attemptUpdateSinceTime()
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case command.GetContainerListenerMsg:
		m, cmd = m.handleContainerListenerMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	// #3: Receive a batch of container deltas and update the tree accordingly
	case command.GetContainerDeltasMsg:
		m, cmd = m.handleContainerDeltasMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	// #4: By now, the entity tree should be populated and the user is oriented on the entities page with the selected
	// (highlighted) entity at the top. From now on, when new entities are added to or removed from the tree, the
	// selected entity will be maintained so that the user doesn't press enter on the wrong entity. If the selected
	// entity is removed, the selection will be moved to the next entity in the tree.
	case message.StartMaintainEntitySelectionMsg:
		if m.pages[page.EntitiesPageType] != nil {
			m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithMaintainSelection(true)
		}
		return m, nil

	// #7: Receive a log scanner for a container. At this point:
	// - the entity representing the container in the tree will still be pending
	// - a goroutine will be running that is scanning for container logs and sending them to the scanner's logs chan
	case command.StartedLogScannerMsg:
		m, cmd = m.handleStartedLogScannerMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	// #8: Receive new logs from a log scanner. Add them to the page log buffer and collect more logs
	case command.GetNewLogsMsg:
		m, cmd = m.handleNewLogsMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	// #9: Time to update the logs page with the buffered logs and empty the buffer. If paused, no-op
	case message.BatchUpdateLogsMsg:
		if len(m.pageLogBuffer) > 0 && !m.pauseState {
			m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithAppendedLogs(m.pageLogBuffer)
			m.pageLogBuffer = nil
		}
		return m, tea.Tick(constants.BatchUpdateLogsInterval, func(t time.Time) tea.Msg { return message.BatchUpdateLogsMsg{} })

	case command.LogScannersStoppedMsg:
		m, cmd = m.handleLogScannersStoppedMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case fileio.SaveCompleteMsg:
		toastMsg := msg.SuccessMessage
		if toastMsg == "" {
			toastMsg = msg.ErrMessage
		}
		newToast := toast.New(toastMsg)
		m.toast = newToast
		cmds = append(cmds, tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return toast.TimeoutMsg{ID: newToast.ID} }))
		return m, tea.Batch(cmds...)

	case command.ContentCopiedToClipboardMsg:
		toastMsg := "Copied to clipboard"
		if msg.Err != nil {
			toastMsg = fmt.Sprintf("Error copying to clipboard: %s", msg.Err.Error())
		}
		newToast := toast.New(toastMsg)
		m.toast = newToast
		cmds = append(cmds, tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return toast.TimeoutMsg{ID: newToast.ID} }))
		return m, tea.Batch(cmds...)

	case message.UpdateSinceTimeTextMsg:
		if m.sinceTime.Time.IsZero() {
			return m, nil
		}
		cmd = tea.Tick(
			m.sinceTime.TimeToNextUpdate(),
			func(t time.Time) tea.Msg { return message.UpdateSinceTimeTextMsg{UUID: m.sinceTime.UUID} },
		)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case toast.TimeoutMsg:
		m.toast, cmd = m.toast.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	m.pages[m.currentPageType], cmd = m.pages[m.currentPageType].Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		errString := wrap.String(m.err.Error(), m.width)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"Error - if this seems wrong, consider opening an issue",
			"https://github.com/robino61/kl/issues/new",
			"",
			"ctrl+c to quit",
			"",
			errString,
		)
	}
	if !m.initialized {
		return ""
	}
	topBar := m.topBar()
	if m.helpText != "" {
		centeredHelp := lipgloss.Place(m.width, m.height-m.topBarHeight, lipgloss.Center, lipgloss.Center, m.helpText)
		return lipgloss.JoinVertical(lipgloss.Left, topBar, centeredHelp)
	}
	if m.prompt.Visible {
		return lipgloss.JoinVertical(lipgloss.Left, topBar, m.prompt.View())
	}
	viewLines := strings.Split(m.topBar(), "\n")
	pageView := m.pages[m.currentPageType].View()
	viewLines = append(viewLines, strings.Split(pageView, "\n")...)
	if toastHeight := m.toast.ViewHeight(); m.toast.Visible && toastHeight > 0 {
		viewLines = viewLines[:len(viewLines)-toastHeight]
		viewLines = append(viewLines, strings.Split(m.toast.View(), "\n")...)
	}
	return strings.Join(viewLines, "\n")
}

func (m Model) topBar() string {
	padding := "   "

	sinceTimeText := fmt.Sprintf("Logs for the Last %s", util.TimeSince(m.sinceTime.Time))
	if m.sinceTime.Time.IsZero() {
		sinceTimeText = "Logs for All Time"
	}

	var numPending, numSelected int
	containerEntities := m.entityTree.GetContainerEntities()
	for _, e := range containerEntities {
		if e.LogScannerPending {
			numPending++
		}
		if e.IsSelected() {
			numSelected++
		}
	}
	left := fmt.Sprintf(
		"kl %s%s%s%s%d/%d/%d Pending/Selected/Total",
		m.config.Version,
		padding,
		sinceTimeText,
		padding,
		numPending,
		numSelected,
		len(containerEntities),
	)
	if m.pauseState {
		left += padding + style.Inverse.Render("[PAUSED]")
	}

	right := fmt.Sprintf("%s to quit / %s for help", m.keyMap.Quit.Help().Key, m.keyMap.Help.Help().Key)
	toJoin := []string{left}
	if len(left)+len(padding)+len(right) < m.width {
		toJoin = append(toJoin, right)
	} else {
		toJoin = append(toJoin, strings.Repeat(" ", len(right)))
	}
	return util.JoinWithEqualSpacing(m.width, toJoin...)
}

// startup, shutdown, & bubble tea builtin messages
// ---
func (m Model) initialize() (Model, tea.Cmd) {
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

	// get kubeconfig, accounting for multiple file paths
	kubeconfigPaths := strings.Split(m.config.KubeConfigPath, string(os.PathListSeparator))
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	dev.Debug(fmt.Sprintf("kubeconfig paths: %v", kubeconfigPaths))
	loadingRules.Precedence = kubeconfigPaths
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
	rawKubeConfig, err := clientConfig.RawConfig()
	if err != nil {
		m.err = err
		return m, nil
	}

	// specified contexts
	contextsString := strings.Trim(strings.TrimSpace(m.config.Contexts), ",")
	var contexts []string
	if len(contextsString) > 0 {
		contexts = strings.Split(contextsString, ",")
	}

	// fall back to current context if exists
	if len(contexts) == 0 && rawKubeConfig.CurrentContext != "" {
		contexts = []string{rawKubeConfig.CurrentContext}
	}

	// validate that at least one context was found
	if len(contexts) == 0 {
		m.err = fmt.Errorf("no contexts specified and no current context found in kubeconfig")
		return m, nil
	}

	// validate that each context exists in the kubeconfig
	for _, c := range contexts {
		if _, exists := rawKubeConfig.Contexts[c]; !exists {
			m.err = fmt.Errorf("context %s not found in kubeconfig", c)
			return m, nil
		}
	}
	dev.Debug(fmt.Sprintf("using contexts %v", contexts))

	// get the cluster name for each context
	var clusters []string
	for _, contextName := range contexts {
		clusterName := rawKubeConfig.Contexts[contextName].Cluster
		clusters = append(clusters, clusterName)
	}

	// enforce that each context has a unique cluster, otherwise it will be undefined what auth/namespace to use for that cluster
	clusterToContext := make(map[string]string)
	for _, contextName := range contexts {
		clusterName := rawKubeConfig.Contexts[contextName].Cluster
		if existingContext, exists := clusterToContext[clusterName]; exists {
			m.err = fmt.Errorf("contexts %s and %s both specify cluster %s - unclear which auth/namespace to use", existingContext, contextName, clusterName)
			return m, nil
		}
		clusterToContext[clusterName] = contextName
	}

	// each context has a cluster, auth info, and an optional namespace
	// - if all namespaces are specified, use all namespaces for all clusters
	// - if specific namespaces are specified, ignore context namespaces and use the specified ones for all clusters
	// - if no namespaces are specified, use the context's namespace if it exists, otherwise use the default namespace
	namespacesString := strings.Trim(strings.TrimSpace(m.config.Namespaces), ",")
	var namespaces []string
	if len(namespacesString) > 0 {
		namespaces = strings.Split(namespacesString, ",")
	}

	// ordered pairing of (cluster name, namespaces)
	// not a map
	var allClusterNamespaces []model.ClusterNamespaces
	for _, cluster := range clusters {
		if m.config.AllNamespaces {
			cn := model.ClusterNamespaces{Cluster: cluster, Namespaces: []string{""}}
			allClusterNamespaces = append(allClusterNamespaces, cn)
		} else if len(namespaces) > 0 {
			cn := model.ClusterNamespaces{Cluster: cluster, Namespaces: namespaces}
			allClusterNamespaces = append(allClusterNamespaces, cn)
		} else {
			contextName := clusterToContext[cluster]
			namespace := rawKubeConfig.Contexts[contextName].Namespace
			if namespace == "" {
				namespace = "default"
			}
			cn := model.ClusterNamespaces{Cluster: cluster, Namespaces: []string{namespace}}
			allClusterNamespaces = append(allClusterNamespaces, cn)
		}
	}
	m.allClusterNamespaces = allClusterNamespaces
	for _, cn := range allClusterNamespaces {
		for _, namespace := range cn.Namespaces {
			dev.Debug(fmt.Sprintf("using cluster '%s' namespace '%s'", cn.Cluster, namespace))
		}
	}

	// create a map of cluster to client set
	clusterToClientSet := make(map[string]*kubernetes.Clientset)
	for _, cluster := range clusters {
		// find the context that refers to the configured cluster
		contextName, exists := clusterToContext[cluster]
		if !exists {
			m.err = fmt.Errorf("no context found for cluster %s in kubeconfig", cluster)
			return m, nil
		}

		// set the current context to the one that refers to the configured cluster
		rawKubeConfig.CurrentContext = contextName
		if err := clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), rawKubeConfig, false); err != nil {
			m.err = fmt.Errorf("failed to set current context for cluster %s: %w", cluster, err)
			return m, nil
		}

		dev.Debug(fmt.Sprintf("using context %s for cluster %s", contextName, cluster))

		// reload the client configuration
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
		config, err := clientConfig.ClientConfig()
		if err != nil {
			m.err = fmt.Errorf("failed to get client config for cluster %s: %w", cluster, err)
			return m, nil
		}

		// create a new clientset for the current context
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			m.err = fmt.Errorf("failed to create clientset for cluster %s: %w", cluster, err)
			return m, nil
		}

		// add the clientset to the map
		clusterToClientSet[cluster] = clientset
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.client = k8s.NewClient(ctx, clusterToClientSet)

	m.entityTree = model.NewEntityTree(m.allClusterNamespaces)

	if m.config.LogsView {
		m.currentPageType = page.LogsPageType
	} else {
		m.currentPageType = page.EntitiesPageType
	}
	m.sinceTime = m.config.SinceTime

	m.pages = make(map[page.Type]page.GenericPage)

	m.topBarHeight = lipgloss.Height(m.topBar())
	contentHeight := m.height - m.topBarHeight
	m.pages[page.EntitiesPageType] = page.NewEntitiesPage(m.keyMap, m.width, contentHeight, m.entityTree)
	m.pages[page.LogsPageType] = page.NewLogsPage(m.keyMap, m.width, contentHeight, m.config.Descending)
	m.pages[page.SingleLogPageType] = page.NewSingleLogPage(m.keyMap, m.width, contentHeight)

	m.initialized = true

	// #1: For each namespace in each cluster, subscribe to pod changes that are mapped into container deltas and
	// returned to the app. This builds the initial state of the entity tree and keeps it in sync with cluster states.
	var cmds []tea.Cmd
	for _, clusterNamespaces := range m.allClusterNamespaces {
		for _, namespace := range clusterNamespaces.Namespaces {
			cmds = append(cmds, command.GetContainerListenerCmd(m.client, clusterNamespaces.Cluster, namespace, m.config.Selectors))
		}
	}

	// start updates of since time
	updateSinceTimeTextCmd := tea.Tick(
		m.sinceTime.TimeToNextUpdate(),
		func(t time.Time) tea.Msg { return message.UpdateSinceTimeTextMsg{UUID: m.sinceTime.UUID} },
	)
	cmds = append(cmds, updateSinceTimeTextCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) handleWindowSizeMsg(width, height int) (Model, tea.Cmd) {
	contentHeight := height - m.topBarHeight
	m.prompt.SetWidthAndHeight(width, contentHeight)
	for k, p := range m.pages {
		m.pages[k] = p.WithDimensions(width, contentHeight)
	}
	return m, nil
}

func (m Model) changeCurrentPage(newPage page.Type) (Model, tea.Cmd) {
	switch newPage {
	case page.EntitiesPageType:
		m.currentPageType = page.EntitiesPageType
	case page.LogsPageType:
		// re-enable stickyness on logs page
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithStickyness()
		m.currentPageType = page.LogsPageType
	case page.SingleLogPageType:
		// cancel stickyness on logs page when moving to single log page, otherwise if selection is on the newest log,
		// selection changes when new logs arrive and is not purely driven by the user
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithNoStickyness()
		m.currentPageType = page.SingleLogPageType
	default:
		m.err = fmt.Errorf("unknown page type %d", newPage)
	}
	return m, nil
}

func (m Model) cleanupCmd() tea.Cmd {
	return func() tea.Msg {
		for _, cl := range m.containerListeners {
			if cl.CleanupFunc != nil {
				cl.CleanupFunc()
			}
		}

		if m.entityTree != nil {
			for _, e := range m.entityTree.GetEntities() {
				if e.IsSelected() {
					e.LogScanner.Cancel()
				}
			}
		}

		if m.cancel != nil {
			m.cancel()
		}
		return message.CleanupCompleteMsg{}
	}
}

// tea.KeyMsg handling
// ---

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("App keyMsg: %v", msg))
	defer dev.Debug("App keyMsg complete")

	var cmd tea.Cmd
	var cmds []tea.Cmd

	// #11: After a long and prosperous interactive log session, user exits the app. Some cleanup is done before exiting
	if key.Matches(msg, m.keyMap.Quit) {
		return m, m.cleanupCmd()
	}

	// ignore key messages other than exit if an error is present
	if m.err != nil {
		return m, nil
	}

	// if help text visible, pressing any key will dismiss it
	if m.helpText != "" {
		m.helpText = ""
		return m, nil
	}

	// adjust for buffered input from held keys, e.g "kk" or "jjj"
	msg.Runes = normalizeRunes(msg)

	// if prompt is visible, only allow prompt actions
	if m.prompt.Visible {
		return m.handlePromptKeyMsg(msg)
	}

	// if current page highjacking input, e.g. editing a focused filter, update current page & return
	if m.pages[m.currentPageType].HighjackingInput() {
		m.pages[m.currentPageType], cmd = m.pages[m.currentPageType].Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// update current page
	m.pages[m.currentPageType], cmd = m.pages[m.currentPageType].Update(msg)
	cmds = append(cmds, cmd)

	// change to selection page
	if key.Matches(msg, m.keyMap.Selection) {
		m, cmd = m.changeCurrentPage(page.EntitiesPageType)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// change to logs page
	if key.Matches(msg, m.keyMap.Logs) {
		m, cmd = m.changeCurrentPage(page.LogsPageType)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// save content of current page
	if key.Matches(msg, m.keyMap.Save) {
		cmds = append(cmds, fileio.GetSaveCommand("", m.pages[m.currentPageType].AllContent()))
		return m, tea.Batch(cmds...)
	}

	// show help
	if key.Matches(msg, m.keyMap.Help) {
		m.helpText = m.pages[m.currentPageType].Help()
		return m, nil
	}

	// entities page specific actions
	if m.currentPageType == page.EntitiesPageType {
		m, cmd = m.handleEntitiesPageKeyMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// logs page specific actions
	if m.currentPageType == page.LogsPageType {
		m, cmd = m.handleLogsPageKeyMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// single log page specific actions
	if m.currentPageType == page.SingleLogPageType {
		m, cmd = m.handleSingleLogPageKeyMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleEntitiesPageKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	// handle pressing enter on selected entity
	if key.Matches(msg, m.keyMap.Enter) {
		selected, selectionActions := m.pages[m.currentPageType].(page.EntityPage).GetSelectionActions()
		if len(selectionActions) > constants.ConfirmSelectionActionsThreshold {
			return m.promptToConfirmSelectionActions(selected, selectionActions)
		} else {
			return m.doSelectionActions(selectionActions)
		}
	}

	// change lookback period for logs
	if key.Matches(msg, m.keyMap.Lookback) {
		return m.changeLookbackPeriod(msg)
	}

	// toggle pause state
	if key.Matches(msg, m.keyMap.TogglePause) {
		m.pauseState = !m.pauseState
	}
	return m, nil
}

func (m Model) promptToConfirmSelectionActions(selected model.Entity, selectionActions map[model.Entity]bool) (Model, tea.Cmd) {
	// display a prompt to confirm selection actions
	// use the terminology select & deselect instead of activate, get log scanner, etc.
	var numToActivate, numToDeactivate int
	for _, getLogScanner := range selectionActions {
		if getLogScanner {
			numToActivate++
		} else {
			numToDeactivate++
		}
	}
	topLine := fmt.Sprintf("Select %d visible containers", numToActivate)
	if numToActivate > 0 && numToDeactivate > 0 {
		topLine = fmt.Sprintf("Select %d & deselect %d visible containers", numToActivate, numToDeactivate)
	} else if numToDeactivate > 0 {
		topLine = fmt.Sprintf("Deselect %d visible containers", numToDeactivate)
	}
	topLine = fmt.Sprintf("%s for %s", topLine, selected.Type())
	bottomLine := fmt.Sprintf("%s?", selected.Container.HumanReadable())
	text := []string{topLine, bottomLine}
	m.prompt = prompt.New(true, m.width, m.height-m.topBarHeight, text)
	m.whenPromptConfirm = func() (Model, tea.Cmd) { return m.doSelectionActions(selectionActions) }
	return m, nil
}

func (m Model) doSelectionActions(selectionActions map[model.Entity]bool) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// #6a: The user presses enter on the entities page. Container log scanners are started or stopped accordingly
	// Note that a log scanner can also be started if the container matches the configured Selectors - see 6b
	for entity, startLogScanner := range selectionActions {
		if startLogScanner {
			m, cmd = m.getStartLogScannerCmd(m.client, entity, m.sinceTime.Time)
			cmds = append(cmds, cmd)
		} else {
			// #10a: The user deselects containers. Either the entity is terminated and is removed from the tree, or we
			// stop the log scanner but keep it in the tree
			// Note that a container delta indicating the container is deleted will perform similar actions - see 10b
			if entity.Terminated {
				// user deselecting a terminated container removes it and its logs
				m.removeLogsForContainer(entity.Container)
				m.entityTree.Remove(entity)
			} else {
				m, cmd = m.stopLogScannerBySelectionActionCmd(entity)
				cmds = append(cmds, cmd)
			}
		}
	}

	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithEntityTree(m.entityTree)
	return m, tea.Batch(cmds...)
}

// getStartLogScannerCmd performs the following:
func (m Model) getStartLogScannerCmd(client k8s.Client, entity model.Entity, sinceTime time.Time) (Model, tea.Cmd) {
	// ensure the entity is a container
	err := entity.AssertIsContainer()
	if err != nil {
		m.err = err
		return m, nil
	}

	// ensure the entity does not already have an active log scanner
	if entity.IsSelected() {
		return m, nil
	}

	// mark the entity as pending in the tree so that the UI can show that and to protect against duplicate scanner requests
	entity.LogScannerPending = true
	m.entityTree.AddOrReplace(entity)

	return m, command.StartLogScannerCmd(client, entity.Container, sinceTime)
}

func (m Model) stopLogScannerBySelectionActionCmd(entity model.Entity) (Model, tea.Cmd) {
	var err error
	if !entity.IsSelected() {
		return m, nil
	}
	m, err = m.withContainerEntityPendingAndBufferedLogsRemoved(entity)
	if err != nil {
		m.err = err
		return m, nil
	}
	return m, command.StopLogScannerCmd(entity, false)
}

func (m Model) withContainerEntityPendingAndBufferedLogsRemoved(entity model.Entity) (Model, error) {
	err := entity.AssertIsContainer()
	if err != nil {
		return m, err
	}
	// mark as pending so that in-flight logs are ignored
	entity.LogScannerPending = true
	m.entityTree.AddOrReplace(entity)
	m.removeContainerLogsFromBuffer(entity.Container)
	return m, nil
}

func (m Model) handleLogsPageKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// change to single log page
	if key.Matches(msg, m.keyMap.Enter) {
		selectedLog := m.pages[page.LogsPageType].(page.LogsPage).GetSelectedLog()
		if selectedLog != nil {
			m, cmd = m.changeCurrentPage(page.SingleLogPageType)
			cmds = append(cmds, cmd)
			m.pages[page.SingleLogPageType] = m.pages[page.SingleLogPageType].(page.SingleLogPage).WithLog(*selectedLog)
		}
		return m, tea.Batch(cmds...)
	}

	// change lookback period for logs
	if key.Matches(msg, m.keyMap.Lookback) {
		return m.changeLookbackPeriod(msg)
	}

	// toggle pause state
	if key.Matches(msg, m.keyMap.TogglePause) {
		m.pauseState = !m.pauseState
	}
	return m, nil
}

func (m Model) handleSingleLogPageKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	// handle clear
	var cmd tea.Cmd
	var cmds []tea.Cmd
	isClear := key.Matches(msg, m.keyMap.Clear)
	notHighjackingInput := !m.pages[m.currentPageType].HighjackingInput()
	noAppliedFilter := !m.pages[m.currentPageType].(page.SingleLogPage).HasAppliedFilter()
	if isClear && notHighjackingInput && noAppliedFilter {
		m, cmd = m.changeCurrentPage(page.LogsPageType)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// handle copy single log content
	if key.Matches(msg, m.keyMap.Copy) {
		return m, command.CopyContentToClipboardCmd(strings.Join(m.pages[page.SingleLogPageType].AllContent(), "\n"))
	}

	// handle cycling through single logs
	if key.Matches(msg, m.keyMap.NextLog) || key.Matches(msg, m.keyMap.PrevLog) {
		if key.Matches(msg, m.keyMap.NextLog) {
			m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).ScrolledDownByOne()
		} else {
			m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).ScrolledUpByOne()
		}
		newLog := m.pages[page.LogsPageType].(page.LogsPage).GetSelectedLog()
		if newLog != nil {
			m.pages[page.SingleLogPageType] = m.pages[page.SingleLogPageType].(page.SingleLogPage).WithLog(*newLog)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handlePromptKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	// escape key cancels prompt
	if key.Matches(msg, m.keyMap.Clear) {
		m.prompt.Visible = false
		m.whenPromptConfirm = nil
		return m, nil
	}

	// enter key confirms prompt and optionally runs whenPromptConfirm function
	if key.Matches(msg, m.keyMap.Enter) {
		if m.prompt.ProceedIsSelected() && m.whenPromptConfirm != nil {
			m, cmd = m.whenPromptConfirm()
		}
		m.prompt.Visible = false
		m.whenPromptConfirm = nil
		return m, cmd
	}

	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

func (m Model) changeLookbackPeriod(msg tea.KeyMsg) (Model, tea.Cmd) {
	// if already a lookback change in flight, no additional ones are allowed
	if m.pendingSinceTime != nil {
		return m, nil
	}

	newLookbackMins := getLookbackMins(msg.String())
	newSinceTimestamp := time.Now().Add(-time.Duration(newLookbackMins) * time.Minute)
	if newLookbackMins == -1 {
		newSinceTimestamp = time.Time{}
	}
	newSinceTime := model.NewSinceTime(newSinceTimestamp, newLookbackMins)

	// 0 always available to "reset from now", otherwise can't change to the same lookback
	if newLookbackMins == 0 || newSinceTime != m.sinceTime {
		m.pendingSinceTime = &newSinceTime
		return m.attemptUpdateSinceTime()
	}

	return m, nil
}

// other
// ---

// #2: Receive a container listener i.e. subscription to container deltas for a cluster and namespace
func (m Model) handleContainerListenerMsg(msg command.GetContainerListenerMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}

	// if a container listener already exists for the cluster and namespace, something has gone wrong
	for _, cl := range m.containerListeners {
		if cl.Cluster == msg.Listener.Cluster && cl.Namespace == msg.Listener.Namespace {
			m.err = fmt.Errorf("container listener already exists for cluster %s and namespace %s", msg.Listener.Cluster, msg.Listener.Namespace)
			return m, nil
		}
	}

	// add the container listener and start collecting container deltas in batches for performance
	m.containerListeners = append(m.containerListeners, msg.Listener)
	cmd = command.GetNextContainerDeltasCmd(m.client, msg.Listener, constants.GetNextContainerDeltasDuration)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) handleContainerDeltasMsg(msg command.GetContainerDeltasMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}

	existingContainerEntities := m.entityTree.GetContainerEntities()

	if len(existingContainerEntities) == 0 && !m.seenFirstContainer {
		cmds = append(cmds, tea.Tick(constants.AttemptMaintainEntitySelectionAfterFirstContainer, func(t time.Time) tea.Msg { return message.StartMaintainEntitySelectionMsg{} }))
		m.seenFirstContainer = true
	}

	for _, delta := range msg.DeltaSet.OrderedDeltas() {
		if delta.ToDelete {
			// #10b: if a container delta indicates the container is deleted, remove it from the tree
			for _, existingContainerEntity := range existingContainerEntities {
				if !existingContainerEntity.Container.Equals(delta.Container) {
					continue
				}

				if !existingContainerEntity.IsSelected() {
					// if the existing container has no active log scanner, remove it from the tree
					m.entityTree.Remove(existingContainerEntity)
					dev.Debug(fmt.Sprintf("model removed container %s, state %s", delta.Container.HumanReadable(), delta.Container.Status.State))
				} else {
					// if there is an active log scanner:
					// - keep the entity in the tree
					// - mark it as terminated
					// - stop its log scanner without marking it inactive or removing the logs
					//   - don't remove the logs as it's useful to have them around for debugging
					// - mark the logs, new, old, and in buffer, as coming from a terminated container
					existingContainerEntity.Terminated = true
					existingContainerEntity.Container.Status = delta.Container.Status
					m.entityTree.AddOrReplace(existingContainerEntity)

					cmds = append(cmds, command.StopLogScannerCmd(existingContainerEntity, true))

					m.markLogsTerminatedForContainer(delta.Container)
					dev.Debug(fmt.Sprintf("model terminated container %s, state %s", delta.Container.HumanReadable(), delta.Container.Status.State))
				}
				break
			}
		} else {
			var logScannerAlreadyActive bool
			for _, existingContainerEntity := range existingContainerEntities {
				if existingContainerEntity.Container.Equals(delta.Container) && existingContainerEntity.IsSelected() {
					// preserve logscanner for existing, selected containers - just update status
					logScannerAlreadyActive = true
					updatedEntity := model.Entity{
						Container:         delta.Container,
						LogScanner:        existingContainerEntity.LogScanner,
						LogScannerPending: existingContainerEntity.LogScannerPending,
					}
					m.entityTree.AddOrReplace(updatedEntity)
					break
				}
			}

			// update container statuses for new and inactive containers
			if !logScannerAlreadyActive {
				newEntity := model.Entity{
					Container: delta.Container,
				}
				m.entityTree.AddOrReplace(newEntity)

				// # 6b: if running container without an existing log scanner is auto-selected, start log scanner
				if delta.Selected && delta.Container.Status.State == model.ContainerRunning {
					m, cmd = m.getStartLogScannerCmd(m.client, newEntity, m.sinceTime.Time)
					cmds = append(cmds, cmd)
				}

			}
			dev.Debug(fmt.Sprintf("model added/updated container %s, state %s", delta.Container.HumanReadable(), delta.Container.Status.State))
		}
	}

	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithEntityTree(m.entityTree)
	cmds = append(cmds, command.GetNextContainerDeltasCmd(m.client, msg.Listener, constants.GetNextContainerDeltasDuration))
	return m, tea.Batch(cmds...)
}

func (m Model) handleStartedLogScannerMsg(msg command.StartedLogScannerMsg) (Model, tea.Cmd) {
	existingContainerEntities := m.entityTree.GetContainerEntities()
	logScannerIsForExistingContainer := false
	for _, existingContainerEntity := range existingContainerEntities {
		if msg.LogScanner.Container.Equals(existingContainerEntity.Container) {
			// whether the log scanner started successfully or not, mark the container as no longer pending
			existingContainerEntity.LogScannerPending = false

			// when log scanners start or attempt to start, the container status is checked and updated
			existingContainerEntity.Container.Status = msg.LogScanner.Container.Status

			// update the entity according to the result of the log scanner start attempt
			if msg.Err == nil {
				// consider log scanner startup errors recoverable, e.g. pressing enter on a terminated container
				existingContainerEntity.LogScanner = &msg.LogScanner
			} else {
				dev.Debug(fmt.Sprintf("log scanner startup failed, container %s: %s", msg.LogScanner.Container.HumanReadable(), msg.Err))
				existingContainerEntity.LogScanner = nil
			}

			// update the entity with the successfully started log scanner
			existingContainerEntity.LogScanner = &msg.LogScanner
			m.entityTree.AddOrReplace(existingContainerEntity)

			logScannerIsForExistingContainer = true
			break
		}
	}

	// if the started log scanner is for a container that no longer exists, clean up
	if !logScannerIsForExistingContainer {
		msg.LogScanner.Cancel()
	}

	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithEntityTree(m.entityTree)
	return m.withUpdatedContainerShortNames(), command.GetNextLogsCmd(msg.LogScanner, constants.SingleContainerLogCollectionDuration)
}

func (m Model) handleLogScannersStoppedMsg(msg command.LogScannersStoppedMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// for each entity with a stopped log scanner, mark the scanner as inactive in the tree
	// if the scanner should be restarted, e.g. to update the lookback time, start a new log scanner
	existingEntities := m.entityTree.GetEntities()
	for _, existingEntity := range existingEntities {
		for _, stoppedContainer := range msg.Containers {
			if existingEntity.Container.Equals(stoppedContainer) {
				if !msg.KeepLogs {
					existingEntity.LogScannerPending = false
					existingEntity.LogScanner = nil
					m.entityTree.AddOrReplace(existingEntity)
				}

				if msg.Restart {
					m, cmd = m.getStartLogScannerCmd(m.client, existingEntity, m.sinceTime.Time)
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	// remove all logs for stopped containers
	for _, stoppedContainer := range msg.Containers {
		if !msg.KeepLogs {
			m.removeLogsForContainer(stoppedContainer)
		}
	}

	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithEntityTree(m.entityTree)
	return m.withUpdatedContainerShortNames(), tea.Batch(cmds...)
}

func (m Model) handleNewLogsMsg(msg command.GetNewLogsMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}

	// ignore logs if logScanner has already been closed
	entity := m.entityTree.GetEntity(msg.LogScanner.Container)
	if entity == nil || !entity.IsSelected() {
		return m, nil
	}

	// ignore logs if its from an old logScanner for a container that has been removed and reactivated
	if !entity.LogScanner.Equals(msg.LogScanner) {
		return m, nil
	}

	// ignore logs if logScanner is Pending (i.e. stopping and has had its logs removed)
	if entity.LogScannerPending {
		return m, nil
	}

	// ignore log if logScanner has completed
	if msg.DoneScanning {
		dev.Debug(fmt.Sprintf("done scanning logs for container %s", msg.LogScanner.Container.HumanReadable()))
		return m, nil
	}

	var err error
	var newLogs []model.PageLog
	for _, log := range msg.NewLogs {
		shortName := ""
		if m.containerToShortName != nil {
			shortName, err = m.containerToShortName(log.Container)
			if err != nil {
				m.err = err
				return m, nil
			}
		}
		localTime := log.Timestamp.Local()
		newLog := model.PageLog{
			Log: log,
			ContainerNames: model.PageLogContainerNames{
				Short: shortName,
				Full:  log.Container.HumanReadable(),
			},
			Timestamps: model.PageLogTimestamps{
				Short: localTime.Format(time.TimeOnly),
				Full:  localTime.Format("2006-01-02T15:04:05.000Z07:00"),
			},
			Terminated: entity.Terminated,
		}
		newLogs = append(newLogs, newLog)
	}

	m.pageLogBuffer = append(m.pageLogBuffer, newLogs...)

	if len(newLogs) > 0 {
		dev.Debug(fmt.Sprintf("done adding %d logs, getting next logs", len(newLogs)))
	}
	return m, command.GetNextLogsCmd(msg.LogScanner, constants.SingleContainerLogCollectionDuration)
}

// attemptUpdateSinceTime checks if there are any pending log scanners, and if not, updates the lookback.
// If there are pending log scanners, there's currently no way to stop or cancel pending log scanners,
// so the lookback change is queued up to be attempted again after a delay.
func (m Model) attemptUpdateSinceTime() (Model, tea.Cmd) {
	if m.entityTree.AnyPendingContainers() {
		if !m.toast.Visible && m.pendingSinceTime != nil {
			m.toast = toast.New(getUpdateSinceTimeText(m.pendingSinceTime.LookbackMins))
		}
		return m, tea.Tick(constants.AttemptUpdateSinceTimeInterval, func(t time.Time) tea.Msg { return message.AttemptUpdateSinceTimeMsg{} })
	}
	return m.doUpdateSinceTime()
}

func (m Model) doUpdateSinceTime() (Model, tea.Cmd) {
	if m.pendingSinceTime == nil {
		return m, nil
	}
	// update lookback and indicate it is updated
	m.sinceTime = *m.pendingSinceTime
	m.toast.Visible = false
	m.pendingSinceTime = nil

	// stop all log scanners and signal to restart them with the new lookback
	var err error
	var logScannersToStopAndRestart []model.LogScanner
	for _, containerEntity := range m.entityTree.GetContainerEntities() {
		// leave selected terminated containers untouched as they can't be restarted and may be useful for debugging
		if containerEntity.IsSelected() && !containerEntity.Terminated {
			m, err = m.withContainerEntityPendingAndBufferedLogsRemoved(containerEntity)
			if err != nil {
				m.err = err
				return m, nil
			}
			logScannersToStopAndRestart = append(logScannersToStopAndRestart, *containerEntity.LogScanner)
		}
	}
	return m, command.StopLogScannersInPrepForNewLookbackCmd(logScannersToStopAndRestart)
}

// withUpdatedContainerShortNames updates the container short names in the entity tree and logs page
// it should be called every time the set of active containers changes
func (m Model) withUpdatedContainerShortNames() Model {
	m.containerToShortName = m.entityTree.ContainerToShortName(constants.MinCharsEachSideShortNames)
	newLogsPage, err := m.pages[page.LogsPageType].(page.LogsPage).WithUpdatedShortNames(m.containerToShortName)
	if err != nil {
		m.err = err
		return m
	}
	m.pages[page.LogsPageType] = newLogsPage
	return m
}

func (m *Model) removeLogsForContainer(container model.Container) {
	m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithLogsRemovedForContainer(container)
	m.removeContainerLogsFromBuffer(container)
}

func (m *Model) removeContainerLogsFromBuffer(container model.Container) {
	bufferedLogs := m.pageLogBuffer
	m.pageLogBuffer = nil
	for _, bufferedLog := range bufferedLogs {
		if !bufferedLog.Log.Container.Equals(container) {
			m.pageLogBuffer = append(m.pageLogBuffer, bufferedLog)
		}
	}
}

func (m *Model) markLogsTerminatedForContainer(container model.Container) {
	m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithLogsTerminatedForContainer(container)
	m.markContainerLogsTerminatedInBuffer(container)
}

func (m *Model) markContainerLogsTerminatedInBuffer(container model.Container) {
	for i := range m.pageLogBuffer {
		if m.pageLogBuffer[i].Log.Container.Equals(container) {
			m.pageLogBuffer[i].Terminated = true
		}
	}
}

// normalizeRunes adjusts for buffered key presses
// under high log/update conditions, bubble tea will buffer key presses, so KeyMsg's arrive as e.g. "jjj" or
// "kk" if the user is holding those keys down. This doesn't seem to happen for up/down keys
func normalizeRunes(msg tea.KeyMsg) []rune {
	if len(msg.Runes) > 1 {
		if strings.Trim(msg.String(), "j") == "" {
			return []rune{'j'}
		}
		if strings.Trim(msg.String(), "k") == "" {
			return []rune{'k'}
		}
	}
	return msg.Runes
}

func getLookbackMins(keyString string) int {
	lookbackInt, err := strconv.Atoi(keyString)
	if err != nil {
		panic(fmt.Sprintf("matched lookback but cant parse to int: %s", keyString))
	}
	newLookbackMins, ok := constants.KeyPressToLookbackMins[lookbackInt]
	if !ok {
		panic(fmt.Sprintf("lookback doesn't internally match num mins: %d", lookbackInt))
	}
	return newLookbackMins
}

func getUpdateSinceTimeText(newLookbackMins int) string {
	if newLookbackMins == 0 {
		return "Changing time range to start from now onwards..."
	}
	if newLookbackMins == -1 {
		return "Changing time range to start from first available logs..."
	}
	return fmt.Sprintf("Changing time range to start from %d minutes ago...", newLookbackMins)
}
