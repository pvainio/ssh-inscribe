package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aakso/ssh-inscribe/pkg/client"
	"github.com/aakso/ssh-inscribe/pkg/logging"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "sshi",
	Short: "sshi connects to ssh-inscribed for SSH certificate generation",
}
var ClientConfig = &client.Config{
	UseAgent: true,
	Timeout:  2 * time.Second,
	Retries:  3,
}
var logLevel = "info"

func rootInit() {
	logging.Defaults.DefaultLevel = logLevel
	if err := logging.Setup(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Hacky way to match flags before this subcommand to allow global flags to be set
// There seems to be no way of doing this in Cobra at the moment
func ignoreFlagsAfter(cmds ...string) {
	ignoreFlags := map[string]bool{}
	for _, v := range cmds {
		ignoreFlags[v] = true
	}
	var cmdIndex int
	for i, arg := range os.Args {
		if ignoreFlags[strings.ToLower(arg)] {
			cmdIndex = i
		}
	}

	// Inject -- after the subcommand to signal Cobra not to try to parse flags
	var args []string
	args = append(args, os.Args[:cmdIndex+1]...)
	args = append(args, "--")
	args = append(args, os.Args[cmdIndex+1:]...)
	if err := RootCmd.ParseFlags(args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Rerun rootinit to re-evaluate flag values
	rootInit()
}

func init() {
	cobra.OnInitialize(rootInit)
	RootCmd.PersistentFlags().StringVar(
		&ClientConfig.URL,
		"url",
		os.Getenv("SSH_INSCRIBE_URL"),
		"URL to ssh-inscribed ($SSH_INSCRIBE_URL)",
	)

	defTimeout := ClientConfig.Timeout
	if expire := os.Getenv("SSH_INSCRIBE_TIMEOUT"); expire != "" {
		defTimeout, _ = time.ParseDuration(expire)
	}
	RootCmd.PersistentFlags().DurationVarP(
		&ClientConfig.Timeout,
		"timeout",
		"t",
		defTimeout,
		"Client timeout ($SSH_INSCRIBE_TIMEOUT)",
	)

	retries := ClientConfig.Retries
	if os.Getenv("SSH_INSCRIBE_RETRIES") != "" {
		retries, _ = strconv.Atoi(os.Getenv("SSH_INSCRIBE_RETRIES"))
	}
	RootCmd.PersistentFlags().IntVar(
		&ClientConfig.Retries,
		"retries",
		retries,
		"Set retry on server failure ($SSH_INSCRIBE_RETRIES)",
	)

	if os.Getenv("SSH_INSCRIBE_DEBUG") != "" {
		ClientConfig.Debug = true
	}
	RootCmd.PersistentFlags().BoolVar(
		&ClientConfig.Debug,
		"debug",
		ClientConfig.Debug,
		"Enable request level debugging (contains sensitive data) ($SSH_INSCRIBE_DEBUG)",
	)

	if os.Getenv("SSH_INSCRIBE_INSECURE") != "" {
		ClientConfig.Insecure = true
	}
	RootCmd.PersistentFlags().BoolVar(
		&ClientConfig.Insecure,
		"insecure",
		ClientConfig.Insecure,
		"Disable TLS validation for the server connection (not recommended) ($SSH_INSCRIBE_INSECURE)",
	)

	if os.Getenv("SSH_INSCRIBE_LOGLEVEL") != "" {
		logLevel = os.Getenv("SSH_INSCRIBE_LOGLEVEL")
	}
	RootCmd.PersistentFlags().StringVar(
		&logLevel,
		"loglevel",
		logLevel,
		"Set logging level ($SSH_INSCRIBE_LOGLEVEL)",
	)

	if os.Getenv("SSH_INSCRIBE_QUIET") != "" {
		ClientConfig.Quiet = true
	}
	RootCmd.PersistentFlags().BoolVarP(
		&ClientConfig.Quiet,
		"quiet",
		"q",
		ClientConfig.Quiet,
		"Set quiet mode ($SSH_INSCRIBE_QUIET)",
	)
}
