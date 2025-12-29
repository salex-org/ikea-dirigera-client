/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Prints the version of the IKEA DIRIGERA CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("CLI %s (commit %s, built %s)\n\n", version, commit, date)

		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			fmt.Printf("No version info for Hub available\n", version, commit, date)
			return nil
		}
		dirigeraClient := getDirigeraClient(usedContext)
		hub, err := dirigeraClient.GetHubStatus()
		if err != nil {
			return fmt.Errorf("failed to get hub status: %w", err)
		}
		fmt.Printf("Versions for Hub from context %s\n", usedContextName)
		firmwareVersion, hasFirmwareVersion := hub.Attributes["firmwareVersion"]
		if hasFirmwareVersion {
			fmt.Printf("Firmware: %s \n", firmwareVersion.(string))
		}
		hardwareVersion, hasHardwareVersion := hub.Attributes["hardwareVersion"]
		if hasHardwareVersion {
			fmt.Printf("Hardware: %s \n", hardwareVersion.(string))
		}
		serialNumber, hasSerialNumber := hub.Attributes["serialNumber"]
		if hasSerialNumber {
			fmt.Printf("Serial number: %s \n", serialNumber.(string))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().StringP("context", "c", "", "Defines the context to use")
}
