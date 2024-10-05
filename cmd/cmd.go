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
	// match when possible https://kubernetes.io/docs/reference/kubectl/ & https://kubernetes.io/docs/reference/kubectl/generated/kubectl/
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
		"help": {
			description: `Print usage`,
		},
		"ic": {
			cliShort:      "",
			cfgFileEnvVar: "ignore-container",
			description:   `Ignore containers matching this regex pattern`,
		},
		"iclust": {
			cliShort:      "",
			cfgFileEnvVar: "ignore-cluster",
			description:   `Ignore containers clusters matching this regex pattern`,
		},
		"ins": {
			cliShort:      "",
			cfgFileEnvVar: "ignore-namespace",
			description:   `Ignore namespaces matching this regex pattern`,
		},
		"iown": {
			cliShort:      "",
			cfgFileEnvVar: "ignore-pod-owner",
			description:   `Ignore pod owners matching this regex pattern`,
		},
		"ipod": {
			cliShort:      "",
			cfgFileEnvVar: "ignore-pod",
			description:   `Ignore pods matching this regex pattern`,
		},
		"kubeconfig": {
			cliShort:      "",
			cfgFileEnvVar: "kubeconfig",
			description:   `Config file path. Defaults to $HOME/.kube/config`,
		},
		"limit": {
			cfgFileEnvVar: "limit",
			description:   `Limit the number of selected containers. Default unlimited`,
			isInt:         true,
			defaultIfInt:  -1,
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
			description:   `Auto-select containers matching this regex pattern`,
		},
		"mclust": {
			cliShort:      "",
			cfgFileEnvVar: "match-cluster",
			description:   `Auto-select clusters matching this regex pattern`,
		},
		"mns": {
			cliShort:      "",
			cfgFileEnvVar: "match-namespace",
			description:   `Auto-select namespaces matching this regex pattern`,
		},
		"mown": {
			cliShort:      "",
			cfgFileEnvVar: "match-pod-owner",
			description:   `Auto-select pod owners matching this regex pattern`,
		},
		"mpod": {
			cliShort:      "",
			cfgFileEnvVar: "match-pod",
			description:   `Auto-select pods matching this regex pattern`,
		},
		"namespace": {
			cliShort:      "n",
			cfgFileEnvVar: "namespace",
			description:   `Namespace(s). Can be comma-separated list. Defaults to current namespace`,
		},
		"pod-owner-types": {
			cfgFileEnvVar: "pod-owner-types",
			description:   `Comma-separated list of pod owner types to include. Defaults to ReplicaSet.`,
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
		"ic",
		"iclust",
		"ins",
		"iown",
		"ipod",
		"kubeconfig",
		"limit",
		"logs-view",
		"mc",
		"mclust",
		"mns",
		"mown",
		"mpod",
		"namespace",
		"pod-owner-types",
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
		fmt.Printf("error on kl startup: %v", err)
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

func getContainerLimit(cmd *cobra.Command) int {
	// -1 indicates no limit
	if !cmd.Flags().Lookup("limit").Changed {
		return -1
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		fmt.Printf("error parsing limit: %v\n", err)
		os.Exit(1)
	}
	if limit < 0 {
		fmt.Println("error: limit must be non-negative")
		os.Exit(1)
	}
	return limit
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

func getIgnoreMatchers(cmd *cobra.Command) model.Matcher {
	ignoreMatchers, err := model.NewMatcher(
		model.NewMatcherArgs{
			Cluster:   cmd.Flags().Lookup("iclust").Value.String(),
			Container: cmd.Flags().Lookup("ic").Value.String(),
			PodOwner:  cmd.Flags().Lookup("iown").Value.String(),
			Namespace: cmd.Flags().Lookup("ins").Value.String(),
			Pod:       cmd.Flags().Lookup("ipod").Value.String(),
		},
	)
	if err != nil {
		fmt.Printf("ignore error: %v\n", err)
		os.Exit(1)
	}
	return *ignoreMatchers
}

func getLogsView(cmd *cobra.Command) bool {
	return cmd.Flags().Lookup("logs-view").Value.String() == "true"
}

func getNamespaces(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("namespace").Value.String()
}

func getPodOwnerTypes(cmd *cobra.Command) []string {
	types := strings.Split(cmd.Flags().Lookup("pod-owner-types").Value.String(), ",")
	if len(types) == 0 || (len(types) == 1 && types[0] == "") {
		return []string{"ReplicaSet"}
	}
	return types
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

func getAutoSelectMatchers(cmd *cobra.Command) model.Matcher {
	autoSelectMatchers, err := model.NewMatcher(
		model.NewMatcherArgs{
			Cluster:   cmd.Flags().Lookup("mclust").Value.String(),
			Namespace: cmd.Flags().Lookup("mns").Value.String(),
			PodOwner:  cmd.Flags().Lookup("mown").Value.String(),
			Pod:       cmd.Flags().Lookup("mpod").Value.String(),
			Container: cmd.Flags().Lookup("mc").Value.String(),
		},
	)
	if err != nil {
		fmt.Printf("auto-select error: %v\n", err)
		os.Exit(1)
	}
	return *autoSelectMatchers
}

func getConfig(cmd *cobra.Command) internal.Config {
	return internal.Config{
		AllNamespaces:  getAllNamespaces(cmd),
		ContainerLimit: getContainerLimit(cmd),
		Contexts:       getKubeContexts(cmd),
		Descending:     getDescending(cmd),
		KubeConfigPath: getKubeConfigPath(cmd),
		LogsView:       getLogsView(cmd),
		Matchers: model.Matchers{
			AutoSelectMatcher: getAutoSelectMatchers(cmd),
			IgnoreMatcher:     getIgnoreMatchers(cmd),
		},
		Namespaces:    getNamespaces(cmd),
		PodOwnerTypes: getPodOwnerTypes(cmd),
		SinceTime:     getSince(cmd),
		Version:       getVersion(),
	}
}

func setup(cmd *cobra.Command) (internal.Model, []tea.ProgramOption) {
	initialModel := internal.InitialModel(getConfig(cmd))
	return initialModel, []tea.ProgramOption{tea.WithAltScreen()}
}
