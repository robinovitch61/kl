package client

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/dev"
	"os"
	"os/exec"
	"path/filepath"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ValidateAuthPlugin checks if the gke-gcloud-auth-plugin is available in PATH or at a custom directory.
// If the authInfo is not for gke-gcloud-auth-plugin, continues. Otherwise, check for the plugin's presence.
func ValidateAuthPlugin(authInfo *clientcmdapi.AuthInfo, contextName, gkeAuthPluginDir string) error {
	if authInfo == nil || authInfo.Exec == nil {
		return nil
	}

	if authInfo.Exec.Command != "gke-gcloud-auth-plugin" {
		return nil
	}

	// Look in system PATH first
	pluginPath, err := exec.LookPath(authInfo.Exec.Command)
	if err == nil {
		dev.Debug(fmt.Sprintf("gke-gcloud-auth-plugin found in PATH at %s for context %s", pluginPath, contextName))
		return nil
	}

	// If not in PATH, check --gke-auth-plugin directory if provided
	if gkeAuthPluginDir != "" {
		customPath := filepath.Join(gkeAuthPluginDir, "gke-gcloud-auth-plugin")
		if _, statErr := os.Stat(customPath); statErr == nil {
			dev.Debug(fmt.Sprintf("gke-gcloud-auth-plugin found at custom path: %s", customPath))

			// Prepend the provided directory to PATH so subprocesses can find the Google Cloud SDK
			currentPath := os.Getenv("PATH")
			newPath := fmt.Sprintf("%s%c%s", gkeAuthPluginDir, os.PathListSeparator, currentPath)
			err := os.Setenv("PATH", newPath)
			if err != nil {
				return fmt.Errorf("failed to update PATH environment variable: %w", err)
			}
			return nil

		}

		// Not found at custom path either
		errorMsg := fmt.Sprintf(
			"gke-gcloud-auth-plugin not found in system PATH or at --gke-auth-plugin location: %s\n%s",
			customPath, authInfo.Exec.InstallHint,
		)

		return fmt.Errorf("%s\nUnderlying error: %w", errorMsg, err)
	}

	// Neither in PATH nor custom path provided
	errorMsg := fmt.Sprintf(
		"gke-gcloud-auth-plugin not found in system PATH for context %s.\n"+
			"You can either:\n"+
			"  - %s and ensure 'google-cloud-sdk/bin' is in your system's PATH.\n"+
			"  - Or provide the plugin directory via the --gke-auth-plugin flag to override the default lookup.\n",
		contextName, authInfo.Exec.InstallHint,
	)

	return fmt.Errorf("%s\nUnderlying error: %w", errorMsg, err)
}
