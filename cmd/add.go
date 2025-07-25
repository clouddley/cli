package cmd

import (
	"os"
	"path/filepath"

	"github.com/clouddley/clouddley/internal/log"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add resources to your system",
	Long:    `Use this command to add various resources to your system.`,
	Example: "clouddley add key",
}

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:     "key",
	Short:   "Add SSH Public Key",
	Long:    `Add Clouddley's SSH public key to your server's authorized_keys file to enable secure communication between your server and the Clouddley Platform.`,
	Example: "clouddley add key",
	Run: func(cmd *cobra.Command, args []string) {
		sshPublicKey := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCYDppnNM+F+GtFWaVJXsvobX/i/2uZuLch9386ZEyVtGE1QmRjRkvwEHwFM23STuzbtmqTrYjEnmv3Xkywk7wE0r+OoJxTBIwJP+scg9rAu//3N6CoAKH0Ra1XgdRj8QzqF/1mm4T/Pxtzz3JSpSKwpzW3GtU4NcHuaPAAHavCpahCnZqpPMU90FgRCS9lSmw0EPQcU8kxxeEpFjifip4JBBx/WQuh/8KkBAX/DnWSAO9ynGzPMvOvWPTtQMi7IA7Y8vRWeThfpC/fnU8Tub+99w5h2Y1TnWtUrM49ZMa9WSLtP/+4xKieQPObq0JuX6itNFuuwbb/WHLgOYeZqQTdSeMc6GlSkqniYiAUAv7olBUERHf7QkD7hPOlaw9S/0MCU8DcuujZG2i6UvIkQ60dikvsX8rCiPvfN4Nw1mWh0a1rf9vUxTyCCb+7hh1iPV6RwMx6T4nBjFNjBglHFkYIE5kevLyX2vREJJen+GfZO2GVcnHaNRHBvXZVVEbwt1xRWhAOS+FFtcKUNV+54JsKTaZUEYvfwe/KNjEeOxucljkiK9IYw0IGXB9dtueOTKcirLhpGE9t6LqDhWE05kr0fl/hmnT/g9fHeZDm4jOF71iHogsrZtU5pH8QtTNhaffMkW4EJc+4W0a+boE+/S5Xracbr7D1WBhGC2epXkUWHw== "clouddley-triggr-public-key"`

		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Error("Error getting your home directory", "error", err)
			os.Exit(1)
		}

		authKeyFile := filepath.Join(homeDir, ".ssh", "authorized_keys")

		file, err := os.OpenFile(authKeyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error("Error opening authorized_keys file", "error", err)
			os.Exit(1)
		}

		defer file.Close()

		// Write the public key to the authorized_keys file
		if _, err := file.WriteString(sshPublicKey + "\n"); err != nil {
			log.Error("Error writing to authorized_keys file", "error", err)
			os.Exit(1)
		}

		log.Info("Clouddley's SSH public key has been added successfully")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(keyCmd)
}
