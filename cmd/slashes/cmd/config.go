package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config is a command for display current configuration",
	Long: `config is a command for display current configuration.

It display configuration merged from config file, environment variables
and command line arguments.`,

	Run: configRun,
}

func configRun(cmd *cobra.Command, args []string) {
	config := viper.AllSettings()
	yaml, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("---")
	fmt.Println(string(yaml))
}

func init() {
	// add config command to root command
	rootCmd.AddCommand(configCmd)
}
