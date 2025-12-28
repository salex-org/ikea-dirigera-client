package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/salex-org/ikea-dirigera-client/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const appName = "ikea-dirigera-cli"

type Context struct {
	AccessToken string `mapstructure:"-" yaml:"-" json:"-"` // Wird aus dem Keyring gelesen
	Address     string `mapstructure:"address" yaml:"address" json:"address"`
	Port        int    `mapstructure:"port" yaml:"port" json:"port"`
	Fingerprint string `mapstructure:"fingerprint" yaml:"fingerprint" json:"fingerprint"`
}

type Config struct {
	CurrentContext string              `mapstructure:"current_context" yaml:"current_context"`
	Contexts       map[string]*Context `mapstructure:"contexts" yaml:"contexts"`
}

var appConfig Config

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ikea",
	Short: "A CLI tool for using the API of an IKEA DIRIGERA Hub",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ikea-dirigera-cli.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ikea-dirigera-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ikea-dirigera-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()

	// Unmarshal the config
	if err := viper.Unmarshal(&appConfig); err != nil {
		fmt.Printf("Fehler beim Unmarshal: %v\n", err)
	}

	// Initialize context map if empty
	if appConfig.Contexts == nil {
		appConfig.Contexts = make(map[string]*Context)
	}

	// Read access tokens from keyring
	for name, context := range appConfig.Contexts {
		token, err := keyring.Get(appName, name)
		if err == nil && token != "" {
			context.AccessToken = token
		}
	}
}

func printOutput(cmd *cobra.Command, data any, printTextOutput func(writer io.Writer)) error {
	output := cmd.Flag("output").Value.String()
	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		_ = enc.Encode(data)
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		_ = enc.Encode(data)
	case "text":
		printTextOutput(os.Stdout)
	default:
		return fmt.Errorf("unknown output format: %s", output)
	}

	return nil
}

func getDirigeraClient(context *Context) client.Client {
	return client.Connect(context.Address, context.Port, &client.Authorization{
		AccessToken:    context.AccessToken,
		TLSFingerprint: context.Fingerprint,
	})
}

func getContext(cmd *cobra.Command) (*Context, string, error) {
	contextName := cmd.Flag("context").Value.String()
	if contextName == "" {
		contextName = appConfig.CurrentContext
	}
	if contextName == "" {
		return nil, contextName, fmt.Errorf("context not set")
	}
	context, found := appConfig.Contexts[contextName]
	if !found {
		return nil, contextName, fmt.Errorf("unknown context: %s", contextName)
	}

	return context, contextName, nil
}
