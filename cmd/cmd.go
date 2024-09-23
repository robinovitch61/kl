package cmd

import (
	"fmt"
	"github.com/carlmjohnson/versioninfo"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// Version is public so users can optionally specify or override the version
	// at build time by passing in ldflags, e.g.
	//   go build -ldflags "-X github.com/robinovitch61/kl/cmd.Version=vX.Y.Z"
	Version = ""
)

type arg struct {
	cliShort, cfgFileEnvVar, description, defaultString string
	isBool, isInt, defaultIfBool                        bool
	defaultIfInt                                        int
}

var (
	rootNameToArg = map[string]arg{
		"all-namespaces": {
			cliShort:      "A",
			cfgFileEnvVar: "all-namespaces",
			description:   `If present, view all namespaces. Overrides other specified namespaces`,
			isBool:        true,
		},
		"context": {
			cliShort:      "",
			cfgFileEnvVar: "context",
			description:   `Context(s). Can be a comma-separated list. Defaults to current context`,
		},
		"desc": {
			cliShort:      "d",
			cfgFileEnvVar: "desc",
			description:   `If present, start with logs in descending order by timestamp. Default false`,
			isBool:        true,
		},
		"extra-owner-refs": {
			cfgFileEnvVar: "extra-owner-refs",
			description:   `Comma-separated list of extra owner ref types to include. Defaults to only ReplicaSet.`,
		},
		"help": {
			description: `Print usage`,
		},
		// generally match https://kubernetes.io/docs/reference/kubectl/ & https://kubernetes.io/docs/reference/kubectl/generated/kubectl/
		"kubeconfig": {
			cliShort:      "",
			cfgFileEnvVar: "kubeconfig",
			description:   `Config file path. Defaults to $HOME/.kube/config`,
		},
		"logs-view": {
			cliShort:      "l",
			cfgFileEnvVar: "logs-view",
			description:   `If present, start with logs view. Default false (selection page)`,
			isBool:        true,
		},
		"mc": {
			cliShort:      "",
			cfgFileEnvVar: "match-container",
			description:   `Auto-select matching containers against this regex pattern`,
		},
		"mclust": {
			cliShort:      "",
			cfgFileEnvVar: "match-cluster",
			description:   `Auto-select matching clusters against this regex pattern`,
		},
		"mdep": {
			cliShort:      "",
			cfgFileEnvVar: "match-deployment",
			description:   `Auto-select matching deployments against this regex pattern`,
		},
		"mns": {
			cliShort:      "",
			cfgFileEnvVar: "match-namespace",
			description:   `Auto-select matching namespaces against this regex pattern`,
		},
		"mpod": {
			cliShort:      "",
			cfgFileEnvVar: "match-pod",
			description:   `Auto-select matching pods against this regex pattern`,
		},
		"namespace": {
			cliShort:      "n",
			cfgFileEnvVar: "namespace",
			description:   `Namespace(s). Can be comma-separated list. Defaults to current namespace`,
		},
		"since": {
			cliShort:      "",
			cfgFileEnvVar: "since",
			description:   `Show logs since startup time minus this duration. E.g. 5s, 2m, 1.5h, 2h45m. Default 1m`,
		},
	}

	description = fmt.Sprintf(`kl %s
Leo Robinovitch <leorobinovitch@gmail.com>

kl is an interactive, cross-cluster, multi-container Kubernetes log viewer

Home page: https://github.com/robinovitch61/kl`,
		getVersion(),
	)

	rootCmd = &cobra.Command{
		Use:   "kl",
		Short: "kl: k8s log viewer",
		Long:  description,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd, rootNameToArg)
		},
		Run:     mainEntrypoint,
		Version: getVersion(),
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// init is called once when the cmd package is loaded
// https://golangdocs.com/init-function-in-golang
func init() {
	cliLong := "help"
	rootCmd.PersistentFlags().BoolP(cliLong, rootNameToArg[cliLong].cliShort, rootNameToArg[cliLong].defaultIfBool, rootNameToArg[cliLong].description)

	for _, cliLong = range []string{
		"all-namespaces",
		"context",
		"desc",
		"extra-owner-refs",
		"kubeconfig",
		"logs-view",
		"mc",
		"mclust",
		"mdep",
		"mns",
		"mpod",
		"namespace",
		"since",
	} {
		c := rootNameToArg[cliLong]
		if c.isBool {
			rootCmd.PersistentFlags().BoolP(cliLong, c.cliShort, c.defaultIfBool, c.description)
		} else if c.isInt {
			rootCmd.PersistentFlags().IntP(cliLong, c.cliShort, c.defaultIfInt, c.description)
		} else {
			rootCmd.PersistentFlags().StringP(cliLong, c.cliShort, c.defaultString, c.description)
		}
		_ = viper.BindPFlag(cliLong, rootCmd.PersistentFlags().Lookup(c.cfgFileEnvVar))
	}
	rootCmd.SetVersionTemplate(`{{printf "kl %s\n" .Version}}`)
	rootCmd.Flags().BoolP("version", "v", false, "Show kl version")
}

func initConfig(cmd *cobra.Command, nameToArg map[string]arg) error {
	// bind viper to env vars
	viper.AutomaticEnv()

	bindFlags(cmd, nameToArg)
	return nil
}

func bindFlags(cmd *cobra.Command, nameToArg map[string]arg) {
	v := viper.GetViper()
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		cliLong := f.Name
		viperName := nameToArg[cliLong].cfgFileEnvVar

		// Apply the viper config value to the flag when the flag is not manually specified
		// and viper has a value from the config file or env var
		if !f.Changed && v.IsSet(viperName) {
			val := v.Get(viperName)
			err := cmd.Flags().Set(cliLong, fmt.Sprintf("%v", val))
			if err != nil {
				fmt.Printf("error setting flag %s: %v\n", cliLong, err)
				os.Exit(1)
			}
		}
	})
}

func mainEntrypoint(cmd *cobra.Command, _ []string) {
	initialModel, options := setup(cmd)
	program := tea.NewProgram(initialModel, options...)

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error on kl startup: %v", err)
		os.Exit(1)
	}
}

func getVersion() string {
	if Version != "" {
		return Version
	}
	return versioninfo.Short()
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // Windows
}

func getAllNamespaces(cmd *cobra.Command) bool {
	return cmd.Flags().Lookup("all-namespaces").Value.String() == "true"
}

func getKubeConfigPath(cmd *cobra.Command) string {
	kubeconfig := cmd.Flags().Lookup("kubeconfig").Value.String()
	if kubeconfig != "" {
		return kubeconfig
	}
	return filepath.Join(homeDir(), ".kube", "config")
}

func getKubeContexts(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("context").Value.String()
}

func getDescending(cmd *cobra.Command) bool {
	return cmd.Flags().Lookup("desc").Value.String() == "true"
}

func getExtraOwnerRefs(cmd *cobra.Command) []string {
	return strings.Split(cmd.Flags().Lookup("extra-owner-refs").Value.String(), ",")
}

func getLogsView(cmd *cobra.Command) bool {
	return cmd.Flags().Lookup("logs-view").Value.String() == "true"
}

func getNamespaces(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("namespace").Value.String()
}

func getSince(cmd *cobra.Command) model.SinceTime {
	duration := cmd.Flags().Lookup("since").Value.String()
	if duration == "" {
		return model.NewSinceTime(
			time.Now().Add(-time.Duration(constants.InitialLookbackMins)*time.Minute),
			constants.InitialLookbackMins,
		)
	}
	d, err := time.ParseDuration(duration)
	if err != nil {
		fmt.Printf("error parsing since: %v\n", err)
		os.Exit(1)
	}
	now := time.Now()
	t := now.Add(-d)
	if t.After(now) {
		fmt.Println("error: since time is in the future")
		os.Exit(1)
	}
	return model.NewSinceTime(t, int(d.Minutes()))
}

func getSelectors(cmd *cobra.Command) model.Selectors {
	selectors, err := model.NewSelectors(
		model.NewSelectorArgs{
			Cluster:    cmd.Flags().Lookup("mclust").Value.String(),
			Namespace:  cmd.Flags().Lookup("mns").Value.String(),
			Deployment: cmd.Flags().Lookup("mdep").Value.String(),
			Pod:        cmd.Flags().Lookup("mpod").Value.String(),
			Container:  cmd.Flags().Lookup("mc").Value.String(),
		},
	)
	if err != nil {
		fmt.Printf("selector error: %v\n", err)
		os.Exit(1)
	}
	return *selectors
}

func getConfig(cmd *cobra.Command) internal.Config {
	return internal.Config{
		AllNamespaces:  getAllNamespaces(cmd),
		Contexts:       getKubeContexts(cmd),
		Descending:     getDescending(cmd),
		ExtraOwnerRefs: getExtraOwnerRefs(cmd),
		KubeConfigPath: getKubeConfigPath(cmd),
		LogsView:       getLogsView(cmd),
		Namespaces:     getNamespaces(cmd),
		SinceTime:      getSince(cmd),
		Selectors:      getSelectors(cmd),
		Version:        getVersion(),
	}
}

func setup(cmd *cobra.Command) (internal.Model, []tea.ProgramOption) {
	initialModel := internal.InitialModel(getConfig(cmd))
	return initialModel, []tea.ProgramOption{tea.WithAltScreen()}
}
