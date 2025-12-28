/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// curlCmd represents the curl command
var curlCmd = &cobra.Command{
	Use:   "curl",
	Short: "Call the specified URL in the IKEA DIRIGERA Hub",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("curl called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(curlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// curlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// curlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
