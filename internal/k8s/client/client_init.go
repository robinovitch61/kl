package client

import (
	"context"
	"fmt"
	"github.com/robinovitch61/kl/internal/model"
	"os"
	"strings"

	"github.com/robinovitch61/kl/internal/dev"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func NewK8sClient(
	ctx context.Context,
	kubeConfigPath string,
	contexts []string,
	namespaces []string,
	useAllNamespaces bool,
) (K8sClient, error) {
	rawKubeConfig, loadingRules, err := getKubeConfig(kubeConfigPath)
	if err != nil {
		return clientImpl{}, err
	}

	if len(contexts) == 0 {
		if rawKubeConfig.CurrentContext != "" {
			dev.Debug(fmt.Sprintf("no contexts specified, using current context %s", rawKubeConfig.CurrentContext))
			contexts = []string{rawKubeConfig.CurrentContext}
		} else {
			return nil, fmt.Errorf("no contexts specified and no current context found in kubeconfig")
		}
	}
	for _, c := range contexts {
		if _, exists := rawKubeConfig.Contexts[c]; !exists {
			return nil, fmt.Errorf("context %s not found in kubeconfig", c)
		}
	}
	dev.Debug(fmt.Sprintf("using contexts %v", contexts))

	clusters := make([]string, len(contexts))
	for i := range contexts {
		clusters[i] = rawKubeConfig.Contexts[contexts[i]].Cluster
	}

	clusterToContext := make(map[string]string)
	for _, contextName := range contexts {
		clusterName := rawKubeConfig.Contexts[contextName].Cluster
		if existingContext, exists := clusterToContext[clusterName]; exists {
			return nil, fmt.Errorf("contexts %s and %s both specify cluster %s - unclear which auth/namespace to use", existingContext, contextName, clusterName)
		}
		clusterToContext[clusterName] = contextName
	}

	var allClusterNamespaces []model.ClusterNamespaces
	for _, cluster := range clusters {
		if useAllNamespaces {
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
	for _, cn := range allClusterNamespaces {
		for _, namespace := range cn.Namespaces {
			dev.Debug(fmt.Sprintf("using cluster '%s' namespace '%s'", cn.Cluster, namespace))
		}
	}

	clusterToClientSet, err := createClientSets(clusters, clusterToContext, loadingRules)
	if err != nil {
		return clientImpl{}, err
	}

	return clientImpl{
		ctx:                  ctx,
		clusterToClientset:   clusterToClientSet,
		allClusterNamespaces: allClusterNamespaces,
	}, nil
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
