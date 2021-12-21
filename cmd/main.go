package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"kube-review/pkg/admission"

	"github.com/spf13/cobra"
)

type parameters struct {
	action string
	as     string
	groups []string
}

var (
	params parameters

	rootCmd = &cobra.Command{
		Use:   "kube-review",
		Short: "create admission review requests from provided kubernetes resources",
		Long: `kube-review is a tool to help create AdmissionReview objects from ordinary Kubernetes resource files. 

This is useful when e.g. writing admission controller policies or offline tests of Kubernetes admission controller 
webhooks`,
		Run: func(cmd *cobra.Command, args []string) {
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
				input, err = ioutil.ReadFile(filename)
				if err != nil {
					log.Fatal(err)
				}
			}

			req, err := admission.AdmissionReviewRequest(input, params.action, params.as, params.groups)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(req))
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
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
