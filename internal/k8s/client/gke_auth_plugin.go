package client

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/dev"
	"os/exec"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ValidateAuthPlugin checks if the `gke-gcloud-auth-plugin` is available in the system PATH.
// If the authInfo is not for `gke-gcloud-auth-plugin`, continues. Otherwise, check for the plugin's presence.
func ValidateAuthPlugin(authInfo *clientcmdapi.AuthInfo, contextName string) error {
	if authInfo == nil || authInfo.Exec == nil {
		return nil
	}

	if authInfo.Exec.Command != "gke-gcloud-auth-plugin" {
		return nil
	}

	// Look in system PATH
	pluginPath, err := exec.LookPath(authInfo.Exec.Command)
	if err == nil {
		dev.Debug(fmt.Sprintf("gke-gcloud-auth-plugin found in PATH at %s for context %s", pluginPath, contextName))
		return nil
	}

	// Not found in PATH, construct error message with installation hint
	errorMsg := fmt.Sprintf(
		"gke-gcloud-auth-plugin not found in system PATH for context %s.\n"+
			"  - %s and ensure 'google-cloud-sdk/bin' is in your system's PATH.\n",
		contextName, authInfo.Exec.InstallHint,
	)

	return fmt.Errorf("%s\nUnderlying error: %w", errorMsg, err)
}
