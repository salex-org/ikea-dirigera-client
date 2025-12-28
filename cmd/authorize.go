/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"strings"

	"github.com/google/uuid"
	"github.com/salex-org/ikea-dirigera-client/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

// authorizeCmd represents the authorize command
var authorizeCmd = &cobra.Command{
	Use: "authorize <ip-address>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if net.ParseIP(args[0]) == nil {
			return fmt.Errorf("invalid IP address: %s", args[0])
		}
		return nil
	},
	Aliases: []string{"auth", "a"},
	Short:   "Authorizes a new user on a IKEA dirigera hub",
	Long: `Creates a new user on a IKEA dirigera hub and creates an access token for the API. Optionally creates a new
context in the CLI for the hub to be used by the user. During the authorization the button on the backside of the hub
needs to be pressed!

Examples:

ikea authorize 192.168.1.1

ikea authorize 192.168.1.1 --port 1234

ikea authorize 192.168.1.1 --context my-context

ikea authorize 192.168.1.1 --no-context`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		port, _ := cmd.Flags().GetInt("port")
		contextName, _ := cmd.Flags().GetString("context")
		skipContext, _ := cmd.Flags().GetBool("no-context")

		context, err := authorize(ip, port)
		if err != nil {
			return fmt.Errorf("authorize failed: %v", err)
		}
		fmt.Printf("Created new user on IKEA DIRIGERA Hub %s\n", context.Address)

		if skipContext {
			fmt.Printf("Access Token: %s\nTLS Fingerprint:%s\n\n", context.AccessToken, context.Fingerprint)
			return nil
		}
		if contextName == "" {
			contextName, err = getHubName(context)
			if err != nil {
				return fmt.Errorf("reading hub name failed: %v", err)
			}
		}

		appConfig.Contexts[contextName] = context
		viper.Set("contexts", appConfig.Contexts)

		if appConfig.CurrentContext == "" {
			appConfig.CurrentContext = contextName
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

		err = keyring.Set(appName, contextName, context.AccessToken)
		if err != nil {
			return fmt.Errorf("error writing token to keyring: %w", err)
		}
		fmt.Printf("Created new context with name %s\n", contextName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(authorizeCmd)
	authorizeCmd.Flags().IntP("port", "p", 8443, "The port used to connect to the IKEA DIRIGERA Hub")
	authorizeCmd.Flags().StringP("context", "c", "", "Specifies the name of the context to create (default name of IKEA DIRIGERA Hub)")
	authorizeCmd.Flags().BoolP("no-context", "n", false, "Don't create a context, just create the user")
}

func authorize(ip string, port int) (*Context, error) {
	clientName := generateClientName()
	fmt.Printf("Adding new user %s to IKEA DIRIGERA Hub at %s:%d\n", clientName, ip, port)
	auth, err := client.Authorize(ip, port, clientName, func() {
		fmt.Printf("Please press the button on the backside of the Hub within 1 minute...")
	}, func() {
		fmt.Printf(".")
	})
	if err != nil {
		fmt.Printf("failed: %v\n", err)

		return nil, fmt.Errorf("error authorizing new client: %w", err)
	}
	fmt.Printf("success\n")

	return &Context{
		AccessToken: auth.AccessToken,
		Fingerprint: auth.TLSFingerprint,
		Address:     ip,
		Port:        port,
	}, nil
}

func generateClientName() string {
	hostname, err := os.Hostname()
	hostname = strings.SplitN(hostname, ".", 2)[0]
	if err != nil {
		return uuid.New().String()
	}
	me, err := user.Current()
	if err != nil {
		return uuid.New().String()
	}

	return fmt.Sprintf("%s@%s", me.Username, hostname)
}

func getHubName(context *Context) (string, error) {
	dirigeraClient := client.Connect(context.Address, context.Port, &client.Authorization{
		AccessToken:    context.AccessToken,
		TLSFingerprint: context.Fingerprint,
	})
	hub, err := dirigeraClient.GetHubStatus()
	if err != nil {
		return "", err
	}
	customNameValue, hasCustomName := hub.Attributes["customName"]
	if hasCustomName {
		if customName, isString := customNameValue.(string); isString {
			return customName, nil
		}
	}

	return "", nil
}
