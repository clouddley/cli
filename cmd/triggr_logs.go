package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var followFlag bool

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:     "logs [SERVICE]",
	Aliases: []string{"log"},
	Short:   "View service logs",
	Long:    `View logs for a Docker service. Use -f flag to follow/tail the logs.

NOTE:
  Execute this command on the same machine or virtual machine
  where your Triggr service is running; otherwise no logs will appear.`,
	Example: `clouddley triggr logs servicename
clouddley triggr logs -f servicename`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := args[0]

		// Check if Docker is accessible
		if err := exec.Command("docker", "info").Run(); err != nil {
			color.Yellow("⚠️  Unable to reach your Triggr service host. Run this command on the machine where the service lives.")
		}

		cmdParts := []string{"docker", "service", "logs"}
		if followFlag {
			cmdParts = append(cmdParts, "-f")
		}
		cmdParts = append(cmdParts, service)

		dockerCmd := exec.Command(cmdParts[0], cmdParts[1:]...)

		stdout, err := dockerCmd.StdoutPipe()
		if err != nil {
			color.Red("Error creating stdout pipe: %v", err)
			os.Exit(1)
		}

		stderr, err := dockerCmd.StderrPipe()
		if err != nil {
			color.Red("Error creating stderr pipe: %v", err)
			os.Exit(1)
		}

		if err := dockerCmd.Start(); err != nil {
			color.Red("Error starting docker command: %v", err)
			os.Exit(1)
		}

		// Handle stdout with pretty formatting
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				formatLogLine(line)
			}
		}()

		// Handle stderr (same formatting as stdout)
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				formatLogLine(line)
			}
		}()

		if err := dockerCmd.Wait(); err != nil {
			color.Red("Error executing docker service logs: %v", err)
			os.Exit(1)
		}
	},
}

func formatLogLine(line string) {
	// Split at the first '|' to separate header from message
	parts := strings.SplitN(line, "|", 2)
	if len(parts) < 2 {
		// Fallback for malformed lines
		fmt.Printf("🔶 %s\n", line)
		return
	}

	header := strings.TrimSpace(parts[0])
	message := strings.TrimSpace(parts[1])

	// Extract service name from header (e.g., "nginx.1.d3x5936fm6rv@docker-desktop" -> "nginx")
	serviceName := "unknown"
	if headerParts := strings.Split(header, "."); len(headerParts) > 0 {
		serviceName = headerParts[0]
	}

	// Determine log level based on message content
	emoji, colorFunc := getLogStyle(message)

	// Format: <emoji> <service> › <message>
	colorFunc("%-2s %-10s › %s\n", emoji, serviceName, message)
}

func getLogStyle(message string) (string, func(format string, a ...interface{})) {
	lowerMsg := strings.ToLower(message)

	switch {
	case strings.Contains(lowerMsg, "[error]") || strings.Contains(lowerMsg, "[crit]") ||
		strings.Contains(lowerMsg, "[fail]") || strings.Contains(lowerMsg, "[fatal]") ||
		strings.Contains(lowerMsg, "error:") || strings.Contains(lowerMsg, "fatal:"):
		return "🔴", color.Red

	case strings.Contains(lowerMsg, "[warn]") || strings.Contains(lowerMsg, "[warning]") ||
		strings.Contains(lowerMsg, "warn:") || strings.Contains(lowerMsg, "warning:"):
		return "⚠️", color.Yellow

	case strings.Contains(lowerMsg, "[notice]") || strings.Contains(lowerMsg, "[info]") ||
		strings.Contains(lowerMsg, "notice:") || strings.Contains(lowerMsg, "info:") ||
		strings.Contains(lowerMsg, "start") || strings.Contains(lowerMsg, "ready") ||
		strings.Contains(lowerMsg, "configuration complete") || strings.Contains(lowerMsg, "enabled"):
		return "🔷", color.Blue

	case strings.Contains(lowerMsg, "[debug]") || strings.Contains(lowerMsg, "debug:"):
		return "🔸", color.Cyan

	default:
		return "🔶", color.Cyan
	}
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow log output")
}
