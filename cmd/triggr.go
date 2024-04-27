package cmd

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"

	"github.com/spf13/cobra"
)

// triggrCmd represents the triggr command
var triggrCmd = &cobra.Command{
	Use:   "triggr",
	Short: "Manage Triggr resources",
	Long: `Use this command to manage Triggr resources on the Clouddley Platform. For example:

	clouddley triggr install`,
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install SSH Public Key",
	Long: `Install triggr's SSH public key on your server to enable secure communication between your server and the Clouddley Platform. For example:

	clouddley triggr install`,
	Run: func(cmd *cobra.Command, args []string) {
		sshPublicKey := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCYDppnNM+F+GtFWaVJXsvobX/i/2uZuLch9386ZEyVtGE1QmRjRkvwEHwFM23STuzbtmqTrYjEnmv3Xkywk7wE0r+OoJxTBIwJP+scg9rAu//3N6CoAKH0Ra1XgdRj8QzqF/1mm4T/Pxtzz3JSpSKwpzW3GtU4NcHuaPAAHavCpahCnZqpPMU90FgRCS9lSmw0EPQcU8kxxeEpFjifip4JBBx/WQuh/8KkBAX/DnWSAO9ynGzPMvOvWPTtQMi7IA7Y8vRWeThfpC/fnU8Tub+99w5h2Y1TnWtUrM49ZMa9WSLtP/+4xKieQPObq0JuX6itNFuuwbb/WHLgOYeZqQTdSeMc6GlSkqniYiAUAv7olBUERHf7QkD7hPOlaw9S/0MCU8DcuujZG2i6UvIkQ60dikvsX8rCiPvfN4Nw1mWh0a1rf9vUxTyCCb+7hh1iPV6RwMx6T4nBjFNjBglHFkYIE5kevLyX2vREJJen+GfZO2GVcnHaNRHBvXZVVEbwt1xRWhAOS+FFtcKUNV+54JsKTaZUEYvfwe/KNjEeOxucljkiK9IYw0IGXB9dtueOTKcirLhpGE9t6LqDhWE05kr0fl/hmnT/g9fHeZDm4jOF71iHogsrZtU5pH8QtTNhaffMkW4EJc+4W0a+boE+/S5Xracbr7D1WBhGC2epXkUWHw== “clouddley-triggr-public-key”`

		homeDir, err := os.UserHomeDir()
		if err != nil {
			color.Red("Error getting your home directory: %v", err)
			os.Exit(1)
		}

		authKeyFile := filepath.Join(homeDir, ".ssh", "authorized_keys")

		file, err := os.OpenFile(authKeyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			color.Red("Error opening authorized_keys file: %v", err)
			os.Exit(1)
		}

		defer file.Close()

		// Write the public key to the authorized_keys file
		if _, err := file.WriteString(sshPublicKey + "\n"); err != nil {
			color.Red("Error writing to authorized_keys file: %v", err)
			os.Exit(1)
		}

		color.Green("Triggr's SSH public key has been installed successfully.")

	},
}

func init() {
	rootCmd.AddCommand(triggrCmd)
	triggrCmd.AddCommand(installCmd)

}


