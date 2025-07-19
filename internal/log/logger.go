package log

import (
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	// Global logger instance
	Logger *log.Logger
)

// InitLogger initializes the global logger with custom styling
func InitLogger() {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "Clouddley",
	})

	// Custom styles using lipgloss
	styles := log.DefaultStyles()
	
	// Success style for info level
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("42")).
		Foreground(lipgloss.Color("0"))

	// Error style
	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("196")).
		Foreground(lipgloss.Color("0"))

	// Warning style
	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("214")).
		Foreground(lipgloss.Color("0"))

	// Debug style
	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("8")).
		Foreground(lipgloss.Color("15"))

	Logger.SetStyles(styles)
	Logger.SetLevel(log.InfoLevel)
}

// Convenience functions for common logging patterns
func Info(msg string, keyvals ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Info(msg, keyvals...)
}

func Error(msg string, keyvals ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Error(msg, keyvals...)
}

func Warn(msg string, keyvals ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Warn(msg, keyvals...)
}

func Debug(msg string, keyvals ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Debug(msg, keyvals...)
}

func Print(msg string) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Print(msg)
}
