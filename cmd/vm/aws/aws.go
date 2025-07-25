package aws

import (
	"github.com/spf13/cobra"
)

// AwsCmd represents the aws command
var AwsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Manage AWS EC2 instances",
	Long:  `Create, list, start, stop, and delete AWS EC2 instances with the Clouddley CLI.`,
	Example: `  clouddley vm aws create    # Create a new AWS instance
  clouddley vm aws list      # List AWS instances
  clouddley vm aws start --id i-1234567890abcdef0
  clouddley vm aws stop --id i-1234567890abcdef0
  clouddley vm aws delete --id i-1234567890abcdef0`,
}

func init() {
	// Enable command suggestions for misspelled commands
	AwsCmd.DisableSuggestions = false
	AwsCmd.SuggestionsMinimumDistance = 2
	
	// Add subcommands
	AwsCmd.AddCommand(createCmd)
	AwsCmd.AddCommand(listCmd)
	AwsCmd.AddCommand(startCmd)
	AwsCmd.AddCommand(stopCmd)
	AwsCmd.AddCommand(deleteCmd)
}
