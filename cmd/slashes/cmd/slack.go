package cmd

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HatsuneMiku3939/slashes/pkg/invoker"
	"github.com/HatsuneMiku3939/slashes/pkg/slack"
	"github.com/HatsuneMiku3939/slashes/server"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// slackCmd represents the slack command
var slackCmd = &cobra.Command{
	Use:   "slack",
	Short: "slack is a command for execute slack slash command request",
	Long:  `slack is a command for execute slack slash command request.`,

	Run: slackRun,
}

func slackRun(cmd *cobra.Command, args []string) {
	// root command flags
	command := viper.GetString("command")
	timeout := viper.GetString("timeout")
	port := viper.GetString("port")
	// slack command flags
	path := viper.GetString("slack.url")
	verifyToken := viper.GetString("slack.verify_token")

	// create server
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		logrus.WithError(err).Fatal("failed to parse timeout")
		return
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	HTTPClient := &http.Client{}
	cmdInvoker := invoker.NewCmdInvoker()

	srv := server.New(port, map[string]server.Handler{
		path: slack.New(
			cmdInvoker, HTTPClient, logger,
			command, timeoutDuration, verifyToken),
	})

	// start server in background
	errs := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errs <- err
		}
	}()

	// wait signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errs:
		logrus.WithError(err).Fatal("failed to start server")
	case s := <-sig:
		logrus.WithField("signal", s).Info("received signal")
	}

	// stop server
	if err := srv.Stop(timeoutDuration); err != nil {
		logrus.WithError(err).Fatal("failed to stop server, force exit")
	}
}

func init() {
	// set slack command flags
	slackCmd.Flags().StringP("url", "u", "/slack", "URL path to listen for slash command requests")
	slackCmd.Flags().StringP("verify-token", "v", "", "slack verification token")

	// bind slack command flags to viper
	if err := viper.BindPFlag("slack.url", slackCmd.Flags().Lookup("url")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("slack.verify_token", slackCmd.Flags().Lookup("verify-token")); err != nil {
		panic(err)
	}

	// add slack command to root command
	rootCmd.AddCommand(slackCmd)
}
