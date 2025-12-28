/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/salex-org/ikea-dirigera-client/pkg/client"
	"github.com/spf13/cobra"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for events in the IKEA DIRIGERA Hub",
	Long:  `Writes events to stdout until stopped by Ctrl-C.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}

		// Notification context for reacting on process termination - used by shutdown function
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// Waiting group used to await finishing the shutdown process when stopping
		var wait sync.WaitGroup

		dirigeraClient := getDirigeraClient(usedContext)
		dirigeraClient.RegisterEventHandler(func(event client.Event) {
			fmt.Printf("Event received: %v\n", event)
		}, "deviceStateChanged")

		// Loop function for event listening
		fmt.Printf("Start listening for events in %s...\n", usedContextName)
		wait.Add(1)
		go func() {
			defer wait.Done()
			err := dirigeraClient.ListenForEvents()
			if err != nil {
				log.Fatal(err)
			}
		}()

		// Shutdown function waiting for the SIGTERM notification to stop event listening
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-ctx.Done()
			fmt.Printf("\n\U0001F6D1 Stop listening for events\n")
			err := dirigeraClient.StopEventListening()
			if err != nil {
				log.Fatal(err)
			}
		}()

		wait.Wait()
		fmt.Printf("\U0001F3C1 Shutdown finished\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
	listenCmd.Flags().StringP("context", "c", "", "Defines the context to use")
}
