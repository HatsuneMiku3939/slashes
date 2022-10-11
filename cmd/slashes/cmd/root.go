package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "slashes",
	Short: "slashes: The simplest way to turn your CLI to slack slash command",
	Long: `slashes is a CLI tool to turn your CLI to slack slash command.

`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// set viper read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("slashes")
	viper.AutomaticEnv()

	// set root command flags
	rootCmd.PersistentFlags().StringP("command", "c", "", "absolute path to the command to be executed")
	rootCmd.PersistentFlags().StringP("timeout", "t", "5s", "timeout for the command to be executed")
	rootCmd.PersistentFlags().StringP("port", "p", ":8080", "port to listen for slash command requests")

	// bind root command flags to viper
	if err := viper.BindPFlag("command", rootCmd.PersistentFlags().Lookup("command")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port")); err != nil {
		panic(err)
	}
}
