package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/anderseknert/kube-review/pkg/admission"
	"github.com/spf13/cobra"
)

type parameters struct {
	action string
	as     string
	groups []string
	indent uint8
}

//nolint:gochecknoglobals
var (
	params  parameters
	version = "" // Set by build command
	rootCmd = &cobra.Command{
		Use:   "kube-review",
		Short: "create admission review requests from provided kubernetes resources",
		Long: `kube-review is a tool to help create AdmissionReview objects from ordinary Kubernetes resource files.

This is useful when e.g. writing admission controller policies or offline tests of Kubernetes admission controller
webhooks`,
	}
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "creates an admission review request from provided kubernetes resource",
		Run: func(_ *cobra.Command, args []string) {
			var filename string
			if len(args) > 0 {
				filename = args[0]
			}

			stat, err := os.Stdin.Stat()
			if err != nil {
				log.Fatal(err)
			}
			var input []byte

			// Read data either from stdin, or from file provided as argument
			if (filename == "" || filename == "-") && (stat.Mode()&os.ModeCharDevice) == 0 {
				input, err = io.ReadAll(os.Stdin)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				if filename == "" {
					log.Fatal("No filename provided and no data to read from stdin")
				}
				input, err = os.ReadFile(filename)
				if err != nil {
					log.Fatal(err)
				}
			}

			req, err := admission.CreateAdmissionReviewRequest(input, params.action, params.as, params.groups, params.indent)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprintln(os.Stdout, string(req))
		},
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version of kube-review",
		Long:  `Print the version of kube-review`,
		Run: func(*cobra.Command, []string) {
			fmt.Fprintln(os.Stdout, version)
			os.Exit(0)
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().StringVar(
		&params.action,
		"action",
		"create",
		"Action to simulate (create | update | delete | connect) (default: create)",
	)
	rootCmd.PersistentFlags().StringVar(
		&params.as,
		"as",
		"kube-review",
		"Name of user",
	)
	rootCmd.PersistentFlags().StringSliceVar(
		&params.groups,
		"as-group",
		[]string{},
		"Group(s) of user (may be repeated) (default: empty)",
	)
	rootCmd.PersistentFlags().Uint8Var(
		&params.indent,
		"indent",
		2,
		"Number of spaces to indent JSON output (default: 2)",
	)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
