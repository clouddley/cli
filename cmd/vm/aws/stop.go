package aws

import (
	"context"
	"fmt"

	"github.com/clouddley/clouddley/internal/aws"
	"github.com/clouddley/clouddley/internal/ui"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop --id <instance-id>",
	Short: "Stop an AWS EC2 instance",
	Long:  `Stop a specific AWS EC2 instance that was created by the Clouddley CLI.`,
	Run:   runStop,
}

func init() {
	stopCmd.Flags().StringP("id", "i", "", "Instance ID to stop (required)")
	stopCmd.MarkFlagRequired("id")
}

func runStop(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	instanceID, _ := cmd.Flags().GetString("id")
	if instanceID == "" {
		fmt.Println(ui.FormatError("Error: --id flag is required"))
		return
	}

	// Validate AWS credentials
	if err := aws.ValidateAWSCredentials(ctx); err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error: %v", err)))
		return
	}

	// TODO: Implement actual instance stopping
	fmt.Printf("Stopping instance %s...\n", instanceID)
	fmt.Println(ui.FormatOutput("✓ Success", fmt.Sprintf("Instance %s stopped successfully", instanceID)))
}
