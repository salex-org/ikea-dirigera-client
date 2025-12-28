/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
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

// showUserCmd represents the show user command
var showUserCmd = &cobra.Command{
	Use:     "user <id>",
	Aliases: []string{"u"},
	Short:   "Show details for the user with the specified id",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		user, err := dirigeraClient.GetUser(userID)
		if err != nil {
			return fmt.Errorf("could not get user: %w", err)
		}
		return printOutput(cmd, user, func(writer io.Writer) {
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "ID: %s\nname: %s\ncreated at: %s\n", user.ID, user.Name, user.CreatedAt)
		})
	},
}

// showDeviceCmd represents the show device command
var showDeviceCmd = &cobra.Command{
	Use:     "device <id>",
	Aliases: []string{"dev", "d"},
	Short:   "Show details for the device with the specified id",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceID := args[0]
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		device, err := dirigeraClient.GetDevice(deviceID)
		if err != nil {
			return fmt.Errorf("could not get device: %w", err)
		}
		return printOutput(cmd, device, func(writer io.Writer) {
			t := table.NewWriter()
			t.SetOutputMirror(writer)
			t.AppendHeader(table.Row{"Name", "Value"})
			for attributeName, attributeValue := range device.Attributes {
				t.AppendRow(table.Row{
					attributeName, attributeValue,
				})
			}
			t.SetStyle(table.StyleDefault)
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "ID: %s\ntype: %s\nsub-type: %s\n", device.ID, device.Type, device.DetailedType)
			_, _ = fmt.Fprintf(writer, "is reachable: %t\ncreated at: %s\nlast seen:: %s\n", device.IsReachable, device.CreatedAt, device.LastSeen)
			_, _ = fmt.Fprintf(writer, "found %d attributes:\n", len(device.Attributes))
			t.Render()
		})
	},
}

// showRoomCmd represents the show room command
var showRoomCmd = &cobra.Command{
	Use:     "room <id>",
	Aliases: []string{"r"},
	Short:   "Show details for the room with the specified id",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roomID := args[0]
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		room, err := dirigeraClient.GetRoom(roomID)
		if err != nil {
			return fmt.Errorf("could not get room: %w", err)
		}
		return printOutput(cmd, room, func(writer io.Writer) {
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "ID: %s\nname: %s\n", room.ID, room.Name)
		})
	},
}

// showSceneCmd represents the show scene command
var showSceneCmd = &cobra.Command{
	Use:     "scene <id>",
	Aliases: []string{"s"},
	Short:   "Show details for the scene with the specified id",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sceneID := args[0]
		usedContext, usedContextName, err := getContext(cmd)
		if err != nil {
			return fmt.Errorf("could not get context: %w", err)
		}
		dirigeraClient := getDirigeraClient(usedContext)
		scene, err := dirigeraClient.GetScene(sceneID)
		if err != nil {
			return fmt.Errorf("could not get scene: %w", err)
		}
		return printOutput(cmd, scene, func(writer io.Writer) {
			triggers := table.NewWriter()
			triggers.SetOutputMirror(writer)
			triggers.AppendHeader(table.Row{"ID", "Type", "Disabled"})
			for _, trigger := range scene.Triggers {
				triggers.AppendRow(table.Row{
					trigger.ID, trigger.Type, trigger.Disabled,
				})
			}
			triggers.SetStyle(table.StyleDefault)
			actions := table.NewWriter()
			actions.SetOutputMirror(writer)
			actions.AppendHeader(table.Row{"ID", "Type", "Device", "Attributes"})
			for _, action := range scene.Actions {
				actions.AppendRow(table.Row{
					action.ID, action.Type, action.DeviceID, action.Attributes,
				})
			}
			actions.SetStyle(table.StyleDefault)
			_, _ = fmt.Fprintf(writer, "using context: %s\n", usedContextName)
			_, _ = fmt.Fprintf(writer, "ID: %s\nname: %s\ntype: %s\ncreated at: %s\n", scene.ID, scene.Info.Name, scene.Type, scene.CreatedAt)
			_, _ = fmt.Fprintf(writer, "found %d triggers:\n", len(scene.Triggers))
			triggers.Render()
			_, _ = fmt.Fprintf(writer, "found %d actions:\n", len(scene.Actions))
			actions.Render()
		})
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.PersistentFlags().StringP("context", "c", "", "Defines the context to use")

	showCmd.AddCommand(showTokenCmd)

	showCmd.AddCommand(showUserCmd)
	showUserCmd.Flags().StringP("output", "o", "text", "Defines the format of the output (text or json)")

	showCmd.AddCommand(showDeviceCmd)
	showDeviceCmd.Flags().StringP("output", "o", "text", "Defines the format of the output (text or json)")

	showCmd.AddCommand(showRoomCmd)
	showRoomCmd.Flags().StringP("output", "o", "text", "Defines the format of the output (text or json)")

	showCmd.AddCommand(showSceneCmd)
	showSceneCmd.Flags().StringP("output", "o", "text", "Defines the format of the output (text or json)")
}
