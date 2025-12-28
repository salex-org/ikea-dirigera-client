/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"s"},
	Short:   "Show details for a specified element in the CLI config or the IKEA DIRIGERA Hub",
}

// showTokenCmd represents the show token command
var showTokenCmd = &cobra.Command{
	Use:     "token",
	Aliases: []string{"t"},
	Short:   "Show the access token for the current or specified context",
	RunE: func(cmd *cobra.Command, args []string) error {
		usedContext, _, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		fmt.Print(usedContext.AccessToken)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.PersistentFlags().StringP("context", "c", "", "Defines the context to use")

	showCmd.AddCommand(showTokenCmd)
}
