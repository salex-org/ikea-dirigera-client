/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "d"},
	Short:   "Remove the specified element from the CLI config or the IKEA DIRIGERA Hub",
}

// deleteContextCmd represents the delete context command
var deleteContextCmd = &cobra.Command{
	Use:     "context <name>",
	Aliases: []string{"ctx", "c"},
	Short:   "Remove the specified context from the CLI config and the related user from the IKEA DIRIGERA Hub",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		contextName := args[0]
		context, found := appConfig.Contexts[contextName]
		if !found {
			return fmt.Errorf("unknown context: %s", contextName)
		}
		dirigeraClient := getDirigeraClient(context)
		currentUser, err := dirigeraClient.GetCurrentUser()
		if err != nil {
			fmt.Printf("warning: could not get current user: %w", err)
		}
		err = dirigeraClient.DeleteUser(currentUser.ID)
		if err != nil {
			fmt.Printf("warning: could not delete user: %w", err)
		}

		delete(appConfig.Contexts, contextName)
		viper.Set("contexts", appConfig.Contexts)

		if appConfig.CurrentContext == contextName {
			appConfig.CurrentContext = ""
			viper.Set("current_context", appConfig.CurrentContext)
		}

		err = viper.WriteConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := viper.SafeWriteConfig(); err != nil {
					return fmt.Errorf("error writing config: %w", err)
				}
			} else {
				return fmt.Errorf("error writing config: %w", err)
			}
		}

		err = keyring.Delete(appName, contextName)
		if err != nil {
			return fmt.Errorf("error removing token from keyring: %w", err)
		}
		fmt.Printf("Deleted context %s and user %s\n", contextName, currentUser.ID)

		return nil
	},
}

// deleteUserCmd represents the delete user command
var deleteUserCmd = &cobra.Command{
	Use:     "user <id>",
	Aliases: []string{"u"},
	Short:   "Remove the specified user from the IKEA DIRIGERA Hub",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		err = dirigeraClient.DeleteUser(userID)
		if err != nil {
			return fmt.Errorf("could not delete user %s in %s: %w", userID, usedContextName, err)
		}
		fmt.Printf("User %s deleted in %s\n", userID, usedContextName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.AddCommand(deleteContextCmd)

	deleteCmd.AddCommand(deleteUserCmd)
	deleteUserCmd.Flags().StringP("context", "c", "", "Defines the context to use")
}
