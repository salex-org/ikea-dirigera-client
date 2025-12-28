/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// curlCmd represents the curl command
var curlCmd = &cobra.Command{
	Use:   "curl",
	Short: "Call the specified URL in the IKEA DIRIGERA Hub",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		u, err := url.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid URL format: %w", err)
		}

		if u.Scheme != "" || u.Host != "" || u.Port() != "" {
			return fmt.Errorf("please provide only a path withour schema, host or port")
		}

		if !strings.HasPrefix(args[0], "/") {
			return fmt.Errorf("path must start with a forward slash '/'")
		}

		if u.Path == "" || u.Path == "/" {
			return fmt.Errorf("path cannot be empty")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		usedContext, _, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		response, err := dirigeraClient.Get(path)
		if err != nil {
			return fmt.Errorf("could not get path: %w", err)
		}
		fmt.Print(response)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(curlCmd)
	curlCmd.PersistentFlags().StringP("context", "c", "", "Defines the context to use")
}
