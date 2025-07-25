package vm

import (
	"github.com/clouddley/clouddley/cmd/vm/aws"
	"github.com/spf13/cobra"
)

// VmCmd represents the vm command
var VmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Manage virtual machines across cloud providers",
	Long:  `Create, list, manage virtual machines on AWS, GCP, and Azure cloud providers.`,
	Example: `  clouddley vm aws create    # Create a new AWS instance
  clouddley vm aws list      # List AWS instances
  clouddley vm aws stop --id i-1234567890abcdef0
  clouddley vm aws delete --id i-1234567890abcdef0`,
}

func init() {
	// Enable command suggestions for misspelled commands
	VmCmd.DisableSuggestions = false
	VmCmd.SuggestionsMinimumDistance = 2
	
	// Add AWS subcommands
	VmCmd.AddCommand(aws.AwsCmd)
}
