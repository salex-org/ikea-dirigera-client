package cmd

import (
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/salex-org/ikea-dirigera-client/pkg/client"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List elements from the CLI config or the IKEA DIRIGERA Hub",
}

var listHubsCmd = &cobra.Command{
	Use:     "hubs",
	Aliases: []string{"hub", "h"},
	Short:   "List all available IKEA DIRIGERA Hubs",
	Long:    `Searches for IKEA DIRIGERA Hubs on the local network using mDNS.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hubs, err := client.Scan()
		if err != nil {
			return fmt.Errorf("could not scan for hubs: %w", err)
		}
		return printOutput(cmd, hubs, func(writer io.Writer) {
			t := table.NewWriter()
			t.SetOutputMirror(writer)
			t.AppendHeader(table.Row{"Hostname", "IP", "Port", "Serial Number", "Firmware Version"})
			for _, hub := range hubs {
				t.AppendRow(table.Row{
					hub.HostName, hub.Address, hub.Port, hub.SerialNumber, hub.FirmwareVersion,
				})
			}
			t.SetStyle(table.StyleDefault)
			t.SetAutoIndex(true)
			_, _ = fmt.Fprintf(writer, "found %d hubs:\n", len(hubs))
			t.Render()
		})
	},
}

var listContextsCmd = &cobra.Command{
	Use:     "contexts",
	Aliases: []string{"context", "ctx", "c"},
	Short:   "List all contexts defined in the CLI config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return printOutput(cmd, appConfig.Contexts, func(writer io.Writer) {
			if len(appConfig.Contexts) == 0 {
				_, _ = fmt.Fprintln(writer, "no contexts defined in the CLI config")
				return
			}
			t := table.NewWriter()
			t.SetOutputMirror(writer)
			t.AppendHeader(table.Row{"Current", "Name", "Address", "Port", "TLS Fingerprint"})
			for name, context := range appConfig.Contexts {
				isCurrentContext := name == appConfig.CurrentContext
				marker := ""
				if isCurrentContext {
					marker = "*"
				}
				t.AppendRow(table.Row{
					marker, name, context.Address, context.Port, context.Fingerprint,
				})
			}
			t.SetStyle(table.StyleDefault)
			t.SetAutoIndex(true)
			_, _ = fmt.Fprintf(writer, "%d contexts are defined:\n", len(appConfig.Contexts))
			t.Render()
		})
	},
}

var listDevicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device", "dev", "d"},
	Short:   "List all devices defined in the IKEA DIRIGERA Hub",
	RunE: func(cmd *cobra.Command, args []string) error {
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		devices, err := dirigeraClient.ListDevices()
		if err != nil {
			return fmt.Errorf("could not list devices: %w", err)
		}
		return printOutput(cmd, devices, func(writer io.Writer) {
			t := table.NewWriter()
			t.SetOutputMirror(writer)
			t.AppendHeader(table.Row{"ID", "Type", "Subtype", "Is Reachable"})
			for _, device := range devices {
				t.AppendRow(table.Row{
					device.ID, device.Type, device.DetailedType, device.IsReachable,
				})
			}
			t.SetStyle(table.StyleDefault)
			t.SetAutoIndex(true)
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "found %d devices:\n", len(devices))
			t.Render()
		})
	},
}

var listUsersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user", "u"},
	Short:   "List all users defined in the IKEA DIRIGERA Hub",
	RunE: func(cmd *cobra.Command, args []string) error {
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		users, err := dirigeraClient.ListUsers()
		if err != nil {
			return fmt.Errorf("could not list users: %w", err)
		}
		currentUser, err := dirigeraClient.GetCurrentUser()
		if err != nil {
			return fmt.Errorf("could not get current user: %w", err)
		}
		return printOutput(cmd, users, func(writer io.Writer) {
			t := table.NewWriter()
			t.SetOutputMirror(writer)
			t.AppendHeader(table.Row{"current", "ID", "Name"})
			for _, user := range users {
				isCurrentUser := user.ID == currentUser.ID
				marker := ""
				if isCurrentUser {
					marker = "*"
				}
				t.AppendRow(table.Row{
					marker, user.ID, user.Name,
				})
			}
			t.SetStyle(table.StyleDefault)
			t.SetAutoIndex(true)
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "found %d users:\n", len(users))
			t.Render()
		})
	},
}

var listRoomsCmd = &cobra.Command{
	Use:     "rooms",
	Aliases: []string{"room", "r"},
	Short:   "List all rooms defined in the IKEA DIRIGERA Hub",
	RunE: func(cmd *cobra.Command, args []string) error {
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		rooms, err := dirigeraClient.ListRooms()
		if err != nil {
			return fmt.Errorf("could not list rooms: %w", err)
		}
		return printOutput(cmd, rooms, func(writer io.Writer) {
			t := table.NewWriter()
			t.SetOutputMirror(writer)
			t.AppendHeader(table.Row{"ID", "Name"})
			for _, room := range rooms {
				t.AppendRow(table.Row{
					room.ID, room.Name,
				})
			}
			t.SetStyle(table.StyleDefault)
			t.SetAutoIndex(true)
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "found %d rooms:\n", len(rooms))
			t.Render()
		})
	},
}

var listScenesCmd = &cobra.Command{
	Use:     "scenes",
	Aliases: []string{"scene", "s"},
	Short:   "List all scenes defined in the IKEA DIRIGERA Hub",
	RunE: func(cmd *cobra.Command, args []string) error {
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		scenes, err := dirigeraClient.ListScenes()
		if err != nil {
			return fmt.Errorf("could not list scenes: %w", err)
		}
		return printOutput(cmd, scenes, func(writer io.Writer) {
			t := table.NewWriter()
			t.SetOutputMirror(writer)
			t.AppendHeader(table.Row{"ID", "Name", "Type"})
			for _, scene := range scenes {
				t.AppendRow(table.Row{
					scene.ID, scene.Info.Name, scene.Type,
				})
			}
			t.SetStyle(table.StyleDefault)
			t.SetAutoIndex(true)
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "found %d scenes:\n", len(scenes))
			t.Render()
		})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringP("output", "o", "text", "Defines the format of the output (text or json)")

	listCmd.AddCommand(listHubsCmd)
	listCmd.AddCommand(listContextsCmd)

	listCmd.AddCommand(listDevicesCmd)
	listDevicesCmd.Flags().StringP("context", "c", "", "Defines the context to use")

	listCmd.AddCommand(listUsersCmd)
	listUsersCmd.Flags().StringP("context", "c", "", "Defines the context to use")

	listCmd.AddCommand(listRoomsCmd)
	listRoomsCmd.Flags().StringP("context", "c", "", "Defines the context to use")

	listCmd.AddCommand(listScenesCmd)
	listScenesCmd.Flags().StringP("context", "c", "", "Defines the context to use")
}
