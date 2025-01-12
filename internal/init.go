package internal

import (
	"context"
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/command"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/k8s"
	"github.com/robinovitch61/kl/internal/message"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/page"
	"github.com/robinovitch61/kl/internal/style"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // register OIDC auth provider
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

func initializedModel(m Model) (Model, tea.Cmd) {
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

	m, err := initializeKubeConfig(m)
	if err != nil {
		m.err = err
		return m, nil
	}

	m = initializePages(m)

	cmds := createInitialCommands(m)

	return m, tea.Batch(cmds...)
}

func initializeKubeConfig(m Model) (Model, error) {
	rawKubeConfig, loadingRules, err := getKubeConfig(m.config.KubeConfigPath)
	if err != nil {
		return m, err
	}

	contexts, err := getContexts(m.config.Contexts, rawKubeConfig)
	if err != nil {
		return m, err
	}
	dev.Debug(fmt.Sprintf("using contexts %v", contexts))

	clusters := getClustersFromContexts(contexts, rawKubeConfig)

	clusterToContext, err := validateUniqueClusters(contexts, clusters, rawKubeConfig)
	if err != nil {
		return m, err
	}

	allClusterNamespaces := buildClusterNamespaces(m, clusters, clusterToContext, rawKubeConfig)
	m.allClusterNamespaces = allClusterNamespaces
	logClusterNamespaces(allClusterNamespaces)

	clusterToClientSet, err := createClientSets(clusters, clusterToContext, loadingRules)
	if err != nil {
		return m, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.client = k8s.NewClient(ctx, clusterToClientSet)
	m.entityTree = model.NewEntityTree(m.allClusterNamespaces)

	m.termStyleData = style.NewTermStyleData()

	return m, nil
}

func getClustersFromContexts(contexts []string, rawKubeConfig api.Config) []string {
	var clusters []string
	for _, contextName := range contexts {
		clusterName := rawKubeConfig.Contexts[contextName].Cluster
		clusters = append(clusters, clusterName)
	}
	return clusters
}

func validateUniqueClusters(contexts []string, clusters []string, rawKubeConfig api.Config) (map[string]string, error) {
	clusterToContext := make(map[string]string)
	for _, contextName := range contexts {
		clusterName := rawKubeConfig.Contexts[contextName].Cluster
		if existingContext, exists := clusterToContext[clusterName]; exists {
			return nil, fmt.Errorf("contexts %s and %s both specify cluster %s - unclear which auth/namespace to use", existingContext, contextName, clusterName)
		}
		clusterToContext[clusterName] = contextName
	}
	return clusterToContext, nil
}

func buildClusterNamespaces(m Model, clusters []string, clusterToContext map[string]string, rawKubeConfig api.Config) []model.ClusterNamespaces {
	namespacesString := strings.Trim(strings.TrimSpace(m.config.Namespaces), ",")
	var namespaces []string
	if len(namespacesString) > 0 {
		namespaces = strings.Split(namespacesString, ",")
	}

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
	return allClusterNamespaces
}

func logClusterNamespaces(allClusterNamespaces []model.ClusterNamespaces) {
	for _, cn := range allClusterNamespaces {
		for _, namespace := range cn.Namespaces {
			dev.Debug(fmt.Sprintf("using cluster '%s' namespace '%s'", cn.Cluster, namespace))
		}
	}
}

func createClientSets(clusters []string, clusterToContext map[string]string, loadingRules *clientcmd.ClientConfigLoadingRules) (map[string]*kubernetes.Clientset, error) {
	clusterToClientSet := make(map[string]*kubernetes.Clientset)
	for _, cluster := range clusters {
		clientset, err := createClientSetForCluster(cluster, clusterToContext, loadingRules)
		if err != nil {
			return nil, err
		}
		clusterToClientSet[cluster] = clientset
	}
	return clusterToClientSet, nil
}

func createClientSetForCluster(cluster string, clusterToContext map[string]string, loadingRules *clientcmd.ClientConfigLoadingRules) (*kubernetes.Clientset, error) {
	contextName, exists := clusterToContext[cluster]
	if !exists {
		return nil, fmt.Errorf("no context found for cluster %s in kubeconfig", cluster)
	}

	// create a config override that sets the current context
	overrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	dev.Debug(fmt.Sprintf("using context %s for cluster %s", contextName, cluster))

	// create client config with the override
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config for cluster %s: %w", cluster, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset for cluster %s: %w", cluster, err)
	}

	return clientset, nil
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
	for _, clusterNamespaces := range m.allClusterNamespaces {
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

// getKubeConfig gets kubeconfig, accounting for multiple file paths
func getKubeConfig(kubeConfigPath string) (api.Config, *clientcmd.ClientConfigLoadingRules, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfigPaths := strings.Split(kubeConfigPath, string(os.PathListSeparator))
	dev.Debug(fmt.Sprintf("kubeconfig paths: %v", kubeconfigPaths))

	loadingRules.Precedence = kubeconfigPaths
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
	rawKubeConfig, err := clientConfig.RawConfig()
	if err != nil {
		return api.Config{}, loadingRules, fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	return rawKubeConfig, loadingRules, nil
}

func getContexts(contextString string, config api.Config) ([]string, error) {
	contextsString := strings.Trim(strings.TrimSpace(contextString), ",")
	var contexts []string
	if len(contextsString) > 0 {
		contexts = strings.Split(contextsString, ",")
	}

	if len(contexts) == 0 && config.CurrentContext != "" {
		contexts = []string{config.CurrentContext}
	}

	if len(contexts) == 0 {
		return nil, fmt.Errorf("no contexts specified and no current context found in kubeconfig")
	}

	for _, c := range contexts {
		if _, exists := config.Contexts[c]; !exists {
			return nil, fmt.Errorf("context %s not found in kubeconfig", c)
		}
	}

	return contexts, nil
}
