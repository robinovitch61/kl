package internal

import (
	"context"
	"fmt"
	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/k8s/entity"
	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/muesli/reflow/wrap"
	"github.com/robinovitch61/kl/internal/color"
	"github.com/robinovitch61/kl/internal/command"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/fileio"
	"github.com/robinovitch61/kl/internal/k8s/client"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/message"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/page"
	"github.com/robinovitch61/kl/internal/prompt"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/toast"
	"github.com/robinovitch61/kl/internal/util"
)

type data struct {
	styles        style.Styles
	termStyleData style.TermStyleData
	topBarHeight  int // assumed constant
}

type state struct {
	width, height      int
	initialized        bool
	stylesLoaded       bool
	gotFirstContainers bool
	seenFirstContainer bool
	fullScreen         bool
	pauseState         bool
	sinceTime          model.SinceTime
	pendingSinceTime   *model.SinceTime
	focusedPageType    page.Type
	rightPageType      page.Type
	helpText           string
	err                error
}

type components struct {
	prompt prompt.Model
	// TODO: move this in to prompt?
	whenPromptConfirm func() (Model, tea.Cmd)
	toast             toast.Model
}

type Model struct {
	config     Config
	keyMap     keymap.KeyMap
	state      state
	data       data
	components components
	pages      map[page.Type]page.GenericPage

	// more complex state
	// TODO: push this down into the logs page model?
	pageLogBuffer []model.PageLog
	entityTree    entity.Tree
	// TODO: put these in entity tree?
	containerToShortName func(container.Container) (k8s_model.ContainerNameAndPrefix, error)
	containerIdToColors  map[string]container.ContainerColors

	k8sClient client.K8sClient

	// containerListeners contains a container update listener for each cluster and namespace combo
	containerListeners []client.ContainerListener

	cancel context.CancelFunc
}

func InitialModel(c Config) Model {
	return Model{
		config: c,
		keyMap: keymap.DefaultKeyMap(),
	}
}

func (m Model) Init() (tea.Model, tea.Cmd) {
	return m, tea.Batch(
		tea.Tick(constants.BatchUpdateLogsInterval, func(t time.Time) tea.Msg { return message.BatchUpdateLogsMsg{} }),
		tea.Tick(constants.CheckStylesLoadedDuration, func(t time.Time) tea.Msg { return message.CheckStylesLoadedMsg{} }),
		tea.RequestForegroundColor,
		tea.RequestBackgroundColor,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dev.DebugUpdateMsg("App", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// handle these regardless of m.state.err
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.Quit) {
			return m, m.cleanupCmd()
		}

	case message.CleanupCompleteMsg:
		return m, tea.Quit
	}

	if m.state.err != nil {
		return m, nil
	}

	// only handle these if m.state.err is nil
	switch msg := msg.(type) {
	case message.ErrMsg:
		m.state.err = msg.Err
		return m, nil

	case tea.BackgroundColorMsg:
		m.data.termStyleData.SetBackground(msg)
		if m.data.termStyleData.IsComplete() {
			m.setStyles(style.NewStyles(m.data.termStyleData))
		}
		return m, nil

	case tea.ForegroundColorMsg:
		m.data.termStyleData.SetForeground(msg)
		if m.data.termStyleData.IsComplete() {
			m.setStyles(style.NewStyles(m.data.termStyleData))
		}
		return m, nil

	case message.CheckStylesLoadedMsg:
		if !m.state.stylesLoaded {
			m.setStyles(style.DefaultStyles)
		}
		return m, nil

	// WindowSizeMsg arrives once on startup, then again every time the window is resized
	case tea.WindowSizeMsg:
		m.state.width, m.state.height = msg.Width, msg.Height
		if !m.state.initialized {
			var err error
			m, cmd, err = initializedModel(m)
			if err != nil {
				m.state.err = err
				return m, nil
			}
			cmds = append(cmds, cmd)
		}
		m.syncDimensions()
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case message.AttemptUpdateSinceTimeMsg:
		m, cmd = m.attemptUpdateSinceTime()
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case command.GetContainerListenerMsg:
		m, cmd = m.handleContainerListenerMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case command.GetContainerDeltasMsg:
		if !m.state.gotFirstContainers {
			m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].WithFocus()
			m.state.gotFirstContainers = true
		}
		m, cmd = m.handleContainerDeltasMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case message.StartMaintainEntitySelectionMsg:
		if m.pages[page.EntitiesPageType] != nil {
			m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithMaintainSelection(true)
		}
		return m, nil

	case command.StartedLogScannerMsg:
		m, cmd = m.handleStartedLogScannerMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case command.GetNewLogsMsg:
		m, cmd = m.handleNewLogsMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case message.BatchUpdateLogsMsg:
		if len(m.pageLogBuffer) > 0 && !m.state.pauseState {
			m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithAppendedLogs(m.pageLogBuffer)
			m.pageLogBuffer = nil
		}
		return m, tea.Tick(constants.BatchUpdateLogsInterval, func(t time.Time) tea.Msg { return message.BatchUpdateLogsMsg{} })

	case command.StoppedLogScannersMsg:
		m, cmd = m.handleStoppedLogScannersMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case fileio.SaveCompleteMsg:
		toastMsg := msg.SuccessMessage
		if toastMsg == "" {
			toastMsg = msg.ErrMessage
		}
		newToast := toast.New(toastMsg)
		m.components.toast = newToast
		cmds = append(cmds, tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return toast.TimeoutMsg{ID: newToast.ID} }))
		return m, tea.Batch(cmds...)

	case command.ContentCopiedToClipboardMsg:
		toastMsg := "Copied to clipboard"
		if msg.Err != nil {
			toastMsg = fmt.Sprintf("Error copying to clipboard: %s", msg.Err.Error())
		}
		newToast := toast.New(toastMsg)
		m.components.toast = newToast
		cmds = append(cmds, tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return toast.TimeoutMsg{ID: newToast.ID} }))
		return m, tea.Batch(cmds...)

	case message.UpdateSinceTimeTextMsg:
		if m.state.sinceTime.Time.IsZero() {
			return m, nil
		}
		cmd = tea.Tick(
			m.state.sinceTime.TimeToNextUpdate(),
			func(t time.Time) tea.Msg { return message.UpdateSinceTimeTextMsg{UUID: m.state.sinceTime.UUID} },
		)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case toast.TimeoutMsg:
		m.components.toast, cmd = m.components.toast.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	if m.pages[m.state.focusedPageType] == nil {
		return m, nil
	}
	m.pages[m.state.focusedPageType], cmd = m.pages[m.state.focusedPageType].Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.state.err != nil {
		errString := wrap.String(m.state.err.Error(), m.state.width)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"Error - if this seems wrong, consider opening an issue",
			"https://github.com/robinovitch61/kl/issues/new",
			"",
			"ctrl+c to quit",
			"",
			errString,
		)
	}

	if !m.state.initialized {
		return ""
	}

	topBar := util.StyleStyledString(m.topBar(), m.data.styles.Lilac)
	if m.state.helpText != "" {
		centeredHelp := lipgloss.Place(m.state.width, m.state.height-m.data.topBarHeight, lipgloss.Center, lipgloss.Center, m.state.helpText)
		return lipgloss.JoinVertical(lipgloss.Left, topBar, centeredHelp)
	}

	if m.components.prompt.Visible {
		return lipgloss.JoinVertical(lipgloss.Left, topBar, m.components.prompt.View())
	}

	viewLines := strings.Split(topBar, "\n")

	var pageView string
	if !m.state.fullScreen && m.state.gotFirstContainers {
		leftPageView := m.data.styles.RightBorder.Render(m.pages[page.EntitiesPageType].View())
		rightPageView := m.pages[m.state.rightPageType].View()
		pageView = lipgloss.JoinHorizontal(lipgloss.Left, leftPageView, rightPageView)
	} else {
		if m.state.focusedPageType == page.EntitiesPageType {
			pageView = m.pages[page.EntitiesPageType].View()
		} else {
			pageView = m.pages[m.state.rightPageType].View()
		}
	}

	viewLines = append(viewLines, strings.Split(pageView, "\n")...)
	if toastHeight := m.components.toast.ViewHeight(); m.components.toast.Visible && toastHeight > 0 {
		viewLines = viewLines[:len(viewLines)-toastHeight]
		viewLines = append(viewLines, strings.Split(m.components.toast.View(), "\n")...)
	}
	return strings.Join(viewLines, "\n")
}

func (m Model) topBar() string {
	padding := "   "

	sinceTimeText := fmt.Sprintf("Logs for the Last %s", util.TimeSince(m.state.sinceTime.Time))
	if m.state.sinceTime.Time.IsZero() {
		sinceTimeText = "Logs for All Time"
	}

	var numPending, numSelected int
	containerEntities := m.entityTree.GetContainerEntities()
	for _, e := range containerEntities {
		if e.State == entity.ScannerStarting || e.State == entity.WantScanning {
			numPending++
		}
		if e.State.MayHaveLogs() {
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
	if m.state.pauseState {
		left += padding + m.data.styles.Inverse.Render("[PAUSED]")
	}

	right := fmt.Sprintf("%s to quit / %s for help", m.keyMap.Quit.Help().Key, m.keyMap.Help.Help().Key)
	toJoin := []string{left}
	if lipgloss.Width(left)+lipgloss.Width(padding)+lipgloss.Width(right) < m.state.width {
		toJoin = append(toJoin, right)
	} else {
		toJoin = append(toJoin, strings.Repeat(" ", len(right)))
	}
	return util.JoinWithEqualSpacing(m.state.width, toJoin...)
}

// startup, shutdown, & bubble tea builtin messages
// ---
func (m Model) syncDimensions() (Model, tea.Cmd) {
	contentHeight := m.state.height - m.data.topBarHeight
	m.components.prompt.SetWidthAndHeight(m.state.width, contentHeight)
	leftWidth := int(math.Round(float64(m.state.width) * constants.LeftPageWidthFraction))
	rightWidth := m.state.width - leftWidth - 1
	if m.state.fullScreen {
		if m.state.focusedPageType == page.EntitiesPageType {
			leftWidth = m.state.width
			rightWidth = 0
		} else {
			leftWidth = 0
			rightWidth = m.state.width
		}
	}
	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].WithDimensions(leftWidth, contentHeight)
	m.pages[page.LogsPageType] = m.pages[page.LogsPageType].WithDimensions(rightWidth, contentHeight)
	m.pages[page.SingleLogPageType] = m.pages[page.SingleLogPageType].WithDimensions(rightWidth, contentHeight)
	return m, nil
}

func (m Model) changeFocusedPage(newPage page.Type) (Model, tea.Cmd) {
	switch newPage {
	case page.EntitiesPageType:
		m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].WithBlur()
		m.state.focusedPageType = page.EntitiesPageType
	case page.LogsPageType:
		m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].WithBlur()
		// re-enable stickiness on logs page
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithStickyness()
		m.state.focusedPageType = page.LogsPageType
	case page.SingleLogPageType:
		// cancel stickiness on logs page when moving to single log page
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithNoStickyness()

		if m.state.focusedPageType != page.LogsPageType {
			m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].WithBlur()
		}

		m.state.focusedPageType = page.SingleLogPageType
	default:
		m.state.err = fmt.Errorf("unknown page type %d", newPage)
	}
	m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].WithFocus()
	m.syncDimensions()
	return m, nil
}

func (m Model) cleanupCmd() tea.Cmd {
	return func() tea.Msg {
		for _, cl := range m.containerListeners {
			if cl.Stop != nil {
				cl.Stop()
			}
		}

		if m.entityTree != nil {
			for _, e := range m.entityTree.GetEntities() {
				if e.LogScanner != nil {
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
	dev.Debug(fmt.Sprintf("App handling keyMsg '%v'", msg))
	defer dev.Debug(fmt.Sprintf("App handling keyMsg '%v' complete", msg))

	var cmd tea.Cmd
	var cmds []tea.Cmd

	if !m.state.initialized {
		return m, nil
	}

	// ignore key messages other than exit if an error is present
	if m.state.err != nil {
		// TODO: everywhere m.state.err is set, should also stop scanners, timed updates, & other cleanup without exiting
		return m, nil
	}

	// if help text visible, pressing any key will dismiss it
	if m.state.helpText != "" {
		m.state.helpText = ""
		return m, nil
	}

	// if prompt is visible, only allow prompt actions
	if m.components.prompt.Visible {
		return m.handlePromptKeyMsg(msg)
	}

	// if current page highjacking input, update current page & return
	if m.pages[m.state.focusedPageType].HighjackingInput() {
		m.pages[m.state.focusedPageType], cmd = m.pages[m.state.focusedPageType].Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// toggle filtering with context
	if key.Matches(msg, m.keyMap.Context) {
		m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].ToggleShowContext()
		return m, tea.Batch(cmds...)
	}

	// store whether a filter is currently applied - helps make decisions later like whether to go back
	// from single log page to logs page or just clear the filter upon user pressing ESC
	hasAppliedFilter := m.pages[m.state.focusedPageType].HasAppliedFilter()

	// update current page with key msg
	m.pages[m.state.focusedPageType], cmd = m.pages[m.state.focusedPageType].Update(msg)
	cmds = append(cmds, cmd)

	// change focus to selection page
	if key.Matches(msg, m.keyMap.Selection) || key.Matches(msg, m.keyMap.SelectionFullScreen) {
		m, cmd = m.changeFocusedPage(page.EntitiesPageType)
		cmds = append(cmds, cmd)
		if key.Matches(msg, m.keyMap.SelectionFullScreen) {
			m.setFullscreen(true)
		}
		return m, tea.Batch(cmds...)
	}

	// change focus to logs/single log page
	if key.Matches(msg, m.keyMap.Logs) || key.Matches(msg, m.keyMap.LogsFullScreen) {
		m, cmd = m.changeFocusedPage(m.state.rightPageType)
		cmds = append(cmds, cmd)
		if key.Matches(msg, m.keyMap.LogsFullScreen) {
			m.setFullscreen(true)
		}
		return m, tea.Batch(cmds...)
	}

	// save content of current page
	if key.Matches(msg, m.keyMap.Save) {
		cmds = append(cmds, fileio.GetSaveCommand("", m.pages[m.state.focusedPageType].ContentForFile()))
		return m, tea.Batch(cmds...)
	}

	// toggle fullscreen for focused page
	if key.Matches(msg, m.keyMap.Fullscreen) {
		m.setFullscreen(!m.state.fullScreen)
		return m, nil
	}

	// show help
	if key.Matches(msg, m.keyMap.Help) {
		m.state.helpText = m.pages[m.state.focusedPageType].Help()
		return m, nil
	}

	// change timestamp format
	if key.Matches(msg, m.keyMap.Timestamps) {
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithNewTimestampFormat()
		return m, nil
	}

	// change container name format
	if key.Matches(msg, m.keyMap.Name) {
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithNewNameFormat()
		return m, nil
	}

	// change log order
	if key.Matches(msg, m.keyMap.ReverseOrder) {
		m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithReversedLogOrder()
		return m, nil
	}

	// entities page specific actions
	if m.state.focusedPageType == page.EntitiesPageType {
		m, cmd = m.handleEntitiesPageKeyMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// logs page specific actions
	if m.state.focusedPageType == page.LogsPageType {
		m, cmd = m.handleLogsPageKeyMsg(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// single log page specific actions
	if m.state.focusedPageType == page.SingleLogPageType {
		m, cmd = m.handleSingleLogPageKeyMsg(msg, hasAppliedFilter)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleEntitiesPageKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	// handle pressing enter on selected entity
	if key.Matches(msg, m.keyMap.Enter) {
		selected, selectionActions := m.pages[m.state.focusedPageType].(page.EntityPage).GetSelectionActions()
		if len(selectionActions) > constants.ConfirmSelectionActionsThreshold {
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
			return m.promptToConfirmSelectionActions(text, selectionActions)
		} else {
			return m.doSelectionActions(selectionActions)
		}
	}

	// handle deselecting all containers
	if key.Matches(msg, m.keyMap.DeselectAll) {
		selectionActions := make(map[entity.Entity]bool)
		containerEntities := m.entityTree.GetContainerEntities()
		for i := range containerEntities {
			if !containerEntities[i].State.ActivatesWhenSelected() {
				selectionActions[containerEntities[i]] = false
			}
		}
		if len(selectionActions) > constants.ConfirmSelectionActionsThreshold {
			text := []string{fmt.Sprintf("Deselect all %d containers?", len(selectionActions))}
			return m.promptToConfirmSelectionActions(text, selectionActions)
		} else {
			return m.doSelectionActions(selectionActions)
		}
	}

	// change since time for logs
	if key.Matches(msg, m.keyMap.SinceTime) {
		return m.changeSinceTime(msg)
	}

	// toggle pause state
	if key.Matches(msg, m.keyMap.TogglePause) {
		m.state.pauseState = !m.state.pauseState
	}
	return m, nil
}

func (m Model) promptToConfirmSelectionActions(text []string, selectionActions map[entity.Entity]bool) (Model, tea.Cmd) {
	m.components.prompt = prompt.New(true, m.state.width, m.state.height-m.data.topBarHeight, text, m.data.styles.Inverse)
	m.components.whenPromptConfirm = func() (Model, tea.Cmd) { return m.doSelectionActions(selectionActions) }
	return m, nil
}

func (m Model) doSelectionActions(selectionActions map[entity.Entity]bool) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	for ent, startLogScanner := range selectionActions {
		if startLogScanner {
			newEntity, newTree, actions := ent.Activate(m.entityTree)
			m.entityTree = newTree
			m, cmd = m.doActions(newEntity, actions)
			cmds = append(cmds, cmd)
		} else {
			newEntity, newTree, actions := ent.Deactivate(m.entityTree)
			m.entityTree = newTree
			m, cmd = m.doActions(newEntity, actions)
			cmds = append(cmds, cmd)
		}
	}

	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithEntityTree(m.entityTree)
	return m, tea.Batch(cmds...)
}

func (m Model) getStartLogScannerCmd(client client.K8sClient, ent entity.Entity, sinceTime time.Time) (Model, tea.Cmd) {
	// ensure the entity is a container
	err := ent.AssertIsContainer()
	if err != nil {
		m.state.err = err
		return m, nil
	}

	// ensure the entity does not already have an active log scanner
	if ent.LogScanner != nil {
		return m, nil
	}

	// check the limit of active log scanners isn't reached
	numPendingOrActive := 0
	for _, ce := range m.entityTree.GetContainerEntities() {
		switch ce.State {
		case entity.WantScanning, entity.ScannerStarting, entity.Scanning, entity.ScannerStopping, entity.Deleted:
			numPendingOrActive++
		default:
		}
	}
	if m.config.ContainerLimit >= 0 && numPendingOrActive >= m.config.ContainerLimit {
		newToast := toast.New(fmt.Sprintf("limit of %d selections reached: run kl with --limit flag to increase", m.config.ContainerLimit))
		m.components.toast = newToast
		return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return toast.TimeoutMsg{ID: newToast.ID} })
	}

	return m, command.StartLogScannerCmd(client, ent.Container, sinceTime)
}

func (m Model) handleLogsPageKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// change to single log page
	if key.Matches(msg, m.keyMap.Enter) {
		selectedLog := m.pages[page.LogsPageType].(page.LogsPage).GetSelectedLog()
		if selectedLog != nil {
			m, cmd = m.changeFocusedPage(page.SingleLogPageType)
			cmds = append(cmds, cmd)
			m.state.rightPageType = page.SingleLogPageType
			m.pages[page.SingleLogPageType] = m.pages[page.SingleLogPageType].(page.SingleLogPage).WithLog(*selectedLog)
		}
		return m, tea.Batch(cmds...)
	}

	// change since time period for logs
	if key.Matches(msg, m.keyMap.SinceTime) {
		return m.changeSinceTime(msg)
	}

	// toggle pause state
	if key.Matches(msg, m.keyMap.TogglePause) {
		m.state.pauseState = !m.state.pauseState
	}
	return m, nil
}

func (m Model) handleSingleLogPageKeyMsg(msg tea.KeyMsg, hasAppliedFilter bool) (Model, tea.Cmd) {
	// handle clear
	var cmd tea.Cmd
	var cmds []tea.Cmd
	isClear := key.Matches(msg, m.keyMap.Clear)
	notHighjackingInput := !m.pages[m.state.focusedPageType].HighjackingInput()
	noAppliedFilter := !hasAppliedFilter
	if isClear && notHighjackingInput && noAppliedFilter {
		m, cmd = m.changeFocusedPage(page.LogsPageType)
		cmds = append(cmds, cmd)
		m.state.rightPageType = page.LogsPageType
		return m, tea.Batch(cmds...)
	}

	// handle copy single log content
	if key.Matches(msg, m.keyMap.Copy) {
		content := m.pages[page.SingleLogPageType].(page.SingleLogPage).ContentForClipboard()
		return m, command.CopyContentToClipboardCmd(strings.Join(content, "\n"))
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
		m.components.prompt.Visible = false
		m.components.whenPromptConfirm = nil
		return m, nil
	}

	// enter key confirms prompt and optionally runs whenPromptConfirm function
	if key.Matches(msg, m.keyMap.Enter) {
		if m.components.prompt.ProceedIsSelected() && m.components.whenPromptConfirm != nil {
			m, cmd = m.components.whenPromptConfirm()
		}
		m.components.prompt.Visible = false
		m.components.whenPromptConfirm = nil
		return m, cmd
	}

	m.components.prompt, cmd = m.components.prompt.Update(msg)
	return m, cmd
}

func (m Model) changeSinceTime(msg tea.KeyMsg) (Model, tea.Cmd) {
	// if already a since time change in flight, no additional ones are allowed
	if m.state.pendingSinceTime != nil {
		return m, nil
	}

	newLookbackMins := getLookbackMins(msg.String())
	newSinceTimestamp := time.Now().Add(-time.Duration(newLookbackMins) * time.Minute)
	if newLookbackMins == -1 {
		newSinceTimestamp = time.Time{}
	}
	newSinceTime := model.NewSinceTime(newSinceTimestamp, newLookbackMins)

	// 0 always available to "reset from now", otherwise can't change to the same since time
	if newLookbackMins == 0 || newSinceTime != m.state.sinceTime {
		m.state.pendingSinceTime = &newSinceTime
		return m.attemptUpdateSinceTime()
	}

	return m, nil
}

// other
// ---

func (m Model) handleContainerListenerMsg(msg command.GetContainerListenerMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if msg.Err != nil {
		m.state.err = msg.Err
		return m, nil
	}

	// if a container listener already exists for the cluster and namespace, something has gone wrong
	for _, cl := range m.containerListeners {
		if cl.Cluster == msg.Listener.Cluster && cl.Namespace == msg.Listener.Namespace {
			m.state.err = fmt.Errorf("container listener already exists for cluster %s and namespace %s", msg.Listener.Cluster, msg.Listener.Namespace)
			return m, nil
		}
	}

	// add the container listener and start collecting container deltas in batches for performance
	m.containerListeners = append(m.containerListeners, msg.Listener)
	cmd = command.GetNextContainerDeltasCmd(msg.Listener, constants.GetNextContainerDeltasDuration)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) handleContainerDeltasMsg(msg command.GetContainerDeltasMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if msg.Err != nil {
		m.state.err = msg.Err
		return m, nil
	}

	existingContainerEntities := m.entityTree.GetContainerEntities()

	if len(existingContainerEntities) == 0 && !m.state.seenFirstContainer {
		m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].WithFocus()
		cmds = append(cmds, tea.Tick(constants.AttemptMaintainEntitySelectionAfterFirstContainer, func(t time.Time) tea.Msg { return message.StartMaintainEntitySelectionMsg{} }))
		m.state.seenFirstContainer = true
	}

	for _, delta := range msg.DeltaSet.OrderedDeltas() {
		// get the existing entity for the container, if it exists
		var existingContainerEntity *entity.Entity
		for _, containerEntity := range existingContainerEntities {
			if containerEntity.Container.Equals(delta.Container) {
				existingContainerEntity = &containerEntity
				break
			}
		}

		if delta.ToDelete {
			if existingContainerEntity != nil {
				ent, newTree, actions := existingContainerEntity.Delete(m.entityTree, delta)
				m.entityTree = newTree
				m, cmd = m.doActions(ent, actions)
				cmds = append(cmds, cmd)
			}
		} else {
			if existingContainerEntity == nil {
				ent := entity.Entity{
					Container: delta.Container,
				}
				newEntity, newTree, actions := ent.Create(m.entityTree, delta)
				m.entityTree = newTree
				m, cmd = m.doActions(newEntity, actions)
				cmds = append(cmds, cmd)
			} else {
				ent, newTree, actions := existingContainerEntity.Update(m.entityTree, delta)
				m.entityTree = newTree
				m, cmd = m.doActions(ent, actions)
				cmds = append(cmds, cmd)
			}
		}
	}

	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithEntityTree(m.entityTree)
	cmds = append(cmds, command.GetNextContainerDeltasCmd(msg.Listener, constants.GetNextContainerDeltasDuration))
	return m, tea.Batch(cmds...)
}

func (m Model) handleStartedLogScannerMsg(msg command.StartedLogScannerMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	existingContainerEntities := m.entityTree.GetContainerEntities()
	var startedContainerEntity *entity.Entity
	for _, containerEntity := range existingContainerEntities {
		if msg.LogScanner.Container.Equals(containerEntity.Container) {
			startedContainerEntity = &containerEntity
			break
		}
	}
	if startedContainerEntity == nil {
		msg.LogScanner.Cancel()
		return m, nil
	}

	ent, newTree, actions := startedContainerEntity.ScannerStarted(m.entityTree, msg.Err, msg.LogScanner)
	m.entityTree = newTree
	m, cmd = m.doActions(ent, actions)
	cmds = append(cmds, cmd)

	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].(page.EntityPage).WithEntityTree(m.entityTree)
	cmds = append(cmds, command.GetNextLogsCmd(msg.LogScanner, constants.SingleContainerLogCollectionDuration))
	return m.withUpdatedContainerShortNames(), tea.Batch(cmds...)
}

func (m Model) handleStoppedLogScannersMsg(msg command.StoppedLogScannersMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// for each entity with a stopped log scanner, mark the scanner as inactive in the tree
	// if the scanner should be restarted, e.g. to update the since time, start a new log scanner
	existingEntities := m.entityTree.GetEntities()
	for _, existingEntity := range existingEntities {
		for _, stoppedContainer := range msg.Containers {
			if existingEntity.Container.Equals(stoppedContainer) {
				ent, newTree, actions := existingEntity.ScannerStopped(m.entityTree)
				m.entityTree = newTree
				m, cmd = m.doActions(ent, actions)
				cmds = append(cmds, cmd)
				if msg.Restart {
					ent, newTree, actions = ent.Activate(m.entityTree)
					m.entityTree = newTree
					m, cmd = m.doActions(ent, actions)
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
		m.state.err = msg.Err
		return m, nil
	}

	// ignore logs if logScanner has already been closed
	ent := m.entityTree.GetEntity(msg.LogScanner.Container)
	if ent == nil || ent.LogScanner == nil {
		return m, nil
	}

	// ignore logs if its from an old logScanner for a container that has been removed and reactivated
	if !ent.LogScanner.Equals(msg.LogScanner) {
		return m, nil
	}

	var err error
	var newLogs []model.PageLog
	for i := range msg.NewLogs {
		shortName := k8s_model.ContainerNameAndPrefix{}
		if m.containerToShortName != nil {
			shortName, err = m.containerToShortName(msg.NewLogs[i].Container)
			if err != nil {
				m.state.err = err
				return m, nil
			}
		}
		fullName := k8s_model.ContainerNameAndPrefix{
			Prefix:        msg.NewLogs[i].Container.IDWithoutContainerName(),
			ContainerName: msg.NewLogs[i].Container.Name,
		}
		var containerColors container.ContainerColors
		if m.containerIdToColors != nil {
			containerColors = m.containerIdToColors[msg.NewLogs[i].Container.ID()]
		}
		newLog := model.PageLog{
			Log:             &msg.NewLogs[i],
			ContainerColors: &containerColors,
			ContainerNames: &model.PageLogContainerNames{
				Short: shortName,
				Full:  fullName,
			},
			Terminated: ent.Container.Status.State == container.ContainerTerminated,
			Styles:     &m.data.styles,
		}
		newLogs = append(newLogs, newLog)
	}

	m.pageLogBuffer = append(m.pageLogBuffer, newLogs...)

	if msg.DoneScanning {
		return m, nil
	}
	return m, command.GetNextLogsCmd(msg.LogScanner, constants.SingleContainerLogCollectionDuration)
}

// attemptUpdateSinceTime checks if there are any pending log scanners, and if not, updates the since time.
// If there are pending log scanners, there's currently no way to stop or cancel pending log scanners,
// so the since time change is queued up to be attempted again after a delay.
func (m Model) attemptUpdateSinceTime() (Model, tea.Cmd) {
	if m.entityTree.AnyScannerStarting() {
		if !m.components.toast.Visible && m.state.pendingSinceTime != nil {
			m.components.toast = toast.New(getUpdateSinceTimeText(m.state.pendingSinceTime.LookbackMins))
		}
		return m, tea.Tick(constants.AttemptUpdateSinceTimeInterval, func(t time.Time) tea.Msg { return message.AttemptUpdateSinceTimeMsg{} })
	}
	return m.doUpdateSinceTime()
}

func (m Model) doUpdateSinceTime() (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	if m.state.pendingSinceTime == nil {
		return m, nil
	}
	// update since time and indicate it is updated
	m.state.sinceTime = *m.state.pendingSinceTime
	m.components.toast.Visible = false
	m.state.pendingSinceTime = nil

	// stop all scanning entities and signal to restart them with the new since time
	var logScannersToStopAndRestart []k8s_log.LogScanner
	for _, containerEntity := range m.entityTree.GetContainerEntities() {
		if containerEntity.State == entity.Scanning {
			ent, newTree, actions := containerEntity.Restart(m.entityTree)
			m.entityTree = newTree
			m, cmd = m.doActions(ent, actions)
			cmds = append(cmds, cmd)
			logScannersToStopAndRestart = append(logScannersToStopAndRestart, *ent.LogScanner)
		}
	}
	// bulk stop log scanners together so they begin restarting one by one only after all have stopped
	cmds = append(cmds, command.StopLogScannersInPrepForNewSinceTimeCmd(logScannersToStopAndRestart))
	return m, tea.Batch(cmds...)
}

func (m Model) doActions(ent entity.Entity, actions []entity.EntityAction) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// ensure actions are unique to avoid duplicate processing
	actionSet := make(map[entity.EntityAction]bool)
	for _, action := range actions {
		if actionSet[action] {
			dev.Debug(fmt.Sprintf("duplicate action detected: %s", action))
		}
		actionSet[action] = true
	}

	for action := range actionSet {
		switch action {
		case entity.StartScanner:
			m, cmd = m.getStartLogScannerCmd(m.k8sClient, ent, m.state.sinceTime.Time)
			cmds = append(cmds, cmd)
		case entity.StopScanner:
			cmds = append(cmds, command.StopLogScannerCmd(ent, false))
		case entity.StopScannerKeepLogs:
			cmds = append(cmds, command.StopLogScannerCmd(ent, true))
		case entity.RemoveEntity:
			m.entityTree.Remove(ent)
		case entity.RemoveLogs:
			m.removeLogsForContainer(ent.Container)
		case entity.MarkLogsTerminated:
			m.markLogsTerminatedForContainer(ent.Container)
		default:
			panic(fmt.Sprintf("unknown entity action: %s", action))
		}
	}
	return m, tea.Batch(cmds...)
}

// withUpdatedContainerShortNames updates the container short names in the entity tree and logs page
// it should be called every time the set of active containers changes
func (m Model) withUpdatedContainerShortNames() Model {
	containers := m.entityTree.GetContainerEntities()
	m.containerIdToColors = make(map[string]container.ContainerColors)
	for _, containerEntity := range containers {
		m.containerIdToColors[containerEntity.Container.ID()] = container.ContainerColors{
			ID:   color.GetColor(containerEntity.Container.ID()),
			Name: color.GetColor(containerEntity.Container.Name),
		}
	}
	m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithContainerColors(m.containerIdToColors)

	m.containerToShortName = m.entityTree.ContainerToShortName(constants.MinCharsEachSideShortNames)
	newLogsPage, err := m.pages[page.LogsPageType].(page.LogsPage).WithUpdatedShortNames(m.containerToShortName)
	if err != nil {
		m.state.err = err
		return m
	}

	err = m.updateShortNamesInBuffer()
	if err != nil {
		m.state.err = err
		return m
	}

	m.pages[page.LogsPageType] = newLogsPage
	return m
}

func (m *Model) updateShortNamesInBuffer() error {
	bufferedLogs := m.pageLogBuffer
	m.pageLogBuffer = nil
	for i := range bufferedLogs {
		short, err := m.containerToShortName(bufferedLogs[i].Log.Container)
		if err != nil {
			return err
		}
		bufferedLogs[i].ContainerNames.Short = short
	}
	m.pageLogBuffer = bufferedLogs
	return nil
}

func (m *Model) removeLogsForContainer(container container.Container) {
	m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithLogsRemovedForContainer(container)
	m.removeContainerLogsFromBuffer(container)
}

func (m *Model) removeContainerLogsFromBuffer(container container.Container) {
	bufferedLogs := m.pageLogBuffer
	m.pageLogBuffer = nil
	for _, bufferedLog := range bufferedLogs {
		if !bufferedLog.Log.Container.Equals(container) {
			m.pageLogBuffer = append(m.pageLogBuffer, bufferedLog)
		}
	}
}

func (m *Model) markLogsTerminatedForContainer(container container.Container) {
	m.pages[page.LogsPageType] = m.pages[page.LogsPageType].(page.LogsPage).WithLogsTerminatedForContainer(container)
	m.markContainerLogsTerminatedInBuffer(container)
}

func (m *Model) markContainerLogsTerminatedInBuffer(container container.Container) {
	for i := range m.pageLogBuffer {
		if m.pageLogBuffer[i].Log.Container.Equals(container) {
			m.pageLogBuffer[i].Terminated = true
		}
	}
}

func (m *Model) setFullscreen(fullscreen bool) {
	m.state.fullScreen = fullscreen
	m.syncDimensions()
}

func (m *Model) setStyles(styles style.Styles) {
	m.data.styles = styles
	m.state.stylesLoaded = true
	m.pages[page.EntitiesPageType] = m.pages[page.EntitiesPageType].WithStyles(styles)
	m.pages[page.LogsPageType] = m.pages[page.LogsPageType].WithStyles(styles)
	m.pages[page.SingleLogPageType] = m.pages[page.SingleLogPageType].WithStyles(styles)
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
