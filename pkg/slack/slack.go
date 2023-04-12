package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HatsuneMiku3939/slashes/pkg/invoker"

	"github.com/labstack/echo/v4"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// Handler is the structure representing the slack slash command handler
type Handler struct {
	// Invoker is the command invoker
	Invoker invoker.Invoker
	// HTTPClient is the http client used to send the message
	HTTPClient *http.Client

	// Command is the filesystem path of the command to execute
	Command string
	// Timeout is the timeout for the command handler
	Timeout time.Duration
	// VerificationToken is the token used to verify the request
	VerificationToken string

	// logger is the logger used to log the events
	logger *logrus.Logger
}

// New returns a new Handler
func New(invoker invoker.Invoker, httpClient *http.Client, logger *logrus.Logger, command string, timeout time.Duration, verificationToken string) *Handler {
	return &Handler{
		Invoker:           invoker,
		HTTPClient:        httpClient,
		Command:           command,
		Timeout:           timeout,
		VerificationToken: verificationToken,

		logger: logger,
	}
}

const (
	// notifyTimeout is the timeout for the notification
	notifyTimeout = 10 * time.Second
)

// Handle is the function that handles the slack slash command
func (h *Handler) Handler() func(c echo.Context) error {
	return func(c echo.Context) error {
		// Parse the request as a slack slash command
		cmd, err := slack.SlashCommandParse(c.Request())
		if err != nil {
			h.logger.WithError(err).Error("Failed to parse the request")
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		// Verify the request
		if valid := cmd.ValidateToken(h.VerificationToken); !valid {
			h.logger.WithField("cmd", cmd).Warn("Invalid token")
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		// handle the command in background after the confirmation message is sent
		defer func() {
			go h.handleCommand(cmd)
		}()

		// sent back a confirmation response
		return c.NoContent(http.StatusOK)
	}
}

// handleCommand is the function that handles the command in background
func (h *Handler) handleCommand(cmd slack.SlashCommand) {
	// notify the user that the command is being handled
	if err := h.notifyStart(cmd); err != nil {
		h.logger.WithError(err).Error("Failed to notify command is being handled")
		return
	}

	// invoke the command
	exitCode, invokeOut, invokeErr := h.invoke(context.Background(), cmd)

	// notify the user that the command is finished
	if err := h.notifyFinish(cmd, exitCode, invokeOut, invokeErr); err != nil {
		h.logger.WithError(err).Error("Failed to notify command is finished")
	}
}

// notifyStart is the function that notifies the user that the command is being handled
func (h *Handler) notifyStart(cmd slack.SlashCommand) error {
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()

	return h.postMessage(ctx, cmd, fmt.Sprintf("Invoke Command with %s timeout...\n$ %s %s", h.Timeout, h.Command, cmd.Text))
}

// notifyFinish is the function that notifies the user that the command is finished
func (h *Handler) notifyFinish(cmd slack.SlashCommand, exitCode int, invokeOut string, invokeErr error) error {
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()

	// notify the user that the command is finished
	exitStatus := fmt.Sprintf("Exit code: %d", exitCode)
	if invokeErr != nil {
		var errMessage string
		switch {
		case ctx.Err() == context.Canceled:
			errMessage = "Command canceled"
		case ctx.Err() == context.DeadlineExceeded:
			errMessage = "Command timed out"
		default:
			errMessage = invokeErr.Error()
		}

		h.logger.WithError(invokeErr).WithField("exitCode", exitCode).Error("Failed to invoke the command")
		message := fmt.Sprintf("%s\n\n%s\n%s", invokeOut, errMessage, exitStatus)
		return h.postMessage(ctx, cmd, message)
	}

	// send the output
	var message string
	if exitCode != 0 {
		h.logger.WithField("exitCode", exitCode).Info("Command exited with non-zero exit code")
		message = fmt.Sprintf("%s\n\n%s", invokeOut, exitStatus)
	} else {
		message = invokeOut
	}

	return h.postMessage(ctx, cmd, message)
}

// invoke is the function that invoke the command
func (h *Handler) invoke(ctx context.Context, cmd slack.SlashCommand) (int, string, error) {
	// create a context with a timeout
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()

	// Parse the command as a command line
	args, err := parseShellwords(cmd.Text)
	if err != nil {
		h.logger.WithField("command", cmd.Text).WithError(err).Error("malformed argument")
		return -1, "", fmt.Errorf("malformed argument: %s %w", cmd.Text, err)
	}

	// invoke the command
	h.logger.WithField("command", h.Command).WithField("args", args).Info("Invoking command")
	return h.Invoker.Invoke(ctx, h.Command, args...)
}

// postMessage is the function that sends a message to the user who sent the command
func (h *Handler) postMessage(ctx context.Context, cmd slack.SlashCommand, text string) error {
	// marshal the message
	message, err := json.Marshal(&slack.Msg{
		Text:         fmt.Sprintf("```\n%s\n```", text),
		ResponseType: slack.ResponseTypeEphemeral,
	})
	if err != nil {
		return err
	}

	// send the message
	req, err := http.NewRequestWithContext(ctx, "POST", cmd.ResponseURL, bytes.NewBuffer(message))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := h.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// drain the body
	if _, err = io.Copy(io.Discard, res.Body); err != nil {
		// ignore the error
		h.logger.WithError(err).Warn("Failed to drain the response body")
	}

	return nil
}

// parseShellwords is the function that parses the command line arguments as shellwords
func parseShellwords(args string) ([]string, error) {
	return shellwords.Parse(args)
}
