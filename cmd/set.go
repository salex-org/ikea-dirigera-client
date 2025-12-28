/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a value in the CLI config or the IKEA DIRIGERA Hub",
}

// setCmd represents the set command
var setContextCmd = &cobra.Command{
	Use:     "context <name>",
	Aliases: []string{"ctx", "c"},
	Short:   "Set the current context",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		newContextName := args[0]
		_, found := appConfig.Contexts[newContextName]
		if !found {
			return fmt.Errorf("context %s not found", newContextName)
		}
		appConfig.CurrentContext = newContextName
		viper.Set("current_context", appConfig.CurrentContext)
		err := viper.WriteConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := viper.SafeWriteConfig(); err != nil {
					return fmt.Errorf("error writing config: %w", err)
				}
			} else {
				return fmt.Errorf("error writing config: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setContextCmd)
}
