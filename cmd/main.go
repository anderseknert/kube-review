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

var (
	action  string
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
					fmt.Println("No filename provided and no data to read from stdin")
					os.Exit(1)
				}
				input, err = ioutil.ReadFile(filename)
				if err != nil {
					log.Fatal(err)
				}
			}

			req, err := admission.AdmissionReviewRequest(input, action)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(req))
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().StringVar(
		&action,
		"action",
		"create",
		"Action to simulate (create | update | delete | connect) (default: create)",
	)
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
