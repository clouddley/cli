package aws

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awsinternal "github.com/clouddley/clouddley/internal/aws"
	"github.com/clouddley/clouddley/internal/log"
	"github.com/clouddley/clouddley/internal/ui"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop --id <instance-id1,instance-id2,...> [--yes]",
	Short:   "Stop one or more AWS EC2 instances",
	Long:    `Stop one or more AWS EC2 instances that were created by the Clouddley CLI. Supports comma-separated instance IDs.`,
	Run:     runStop,
}

func init() {
	stopCmd.Flags().StringP("id", "i", "", "Instance ID(s) to stop - supports comma-separated list (required)")
	stopCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	stopCmd.MarkFlagRequired("id")
}

func runStop(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	instanceIDsFlag, _ := cmd.Flags().GetString("id")
	skipConfirmation, _ := cmd.Flags().GetBool("yes")

	if instanceIDsFlag == "" {
		fmt.Println(ui.FormatError("Error: --id flag is required"))
		return
	}

	// Parse instance IDs (support comma-separated list)
	parts := strings.Split(instanceIDsFlag, ",")
	var instanceIDs []string
	for _, id := range parts {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" { // Filter out empty strings
			instanceIDs = append(instanceIDs, trimmed)
		}
	}

	if len(instanceIDs) == 0 {
		fmt.Println(ui.FormatError("Error: No valid instance IDs provided"))
		return
	}

	// Validate AWS credentials
	if err := awsinternal.ValidateAWSCredentials(ctx); err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error: %v", err)))
		return
	}

	// Confirmation prompt (unless --yes flag is used)
	if !skipConfirmation {
		if len(instanceIDs) == 1 {
			fmt.Printf("Are you sure you want to stop instance %s? (y/n): ", instanceIDs[0])
		} else {
			fmt.Printf("Are you sure you want to stop %d instances (%s)? (y/n): ", 
				len(instanceIDs), strings.Join(instanceIDs, ", "))
		}
		
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(ui.FormatError(fmt.Sprintf("Error reading input: %v", err)))
			return
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Operation cancelled")
			return
		}
	}

	// Get EC2 client
	client, err := awsinternal.GetEC2Client(ctx)
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Failed to create EC2 client: %v", err)))
		return
	}

	// Track results
	var successful []string
	var failed []string

	// Stop instances
	for _, instanceID := range instanceIDs {
		fmt.Printf("Stopping instance %s...\n", instanceID)
		
		if err := stopInstance(ctx, client, instanceID); err != nil {
			log.Error("Failed to stop instance", "instance", instanceID, "error", err)
			fmt.Println(ui.FormatError(fmt.Sprintf("Failed to stop %s: %v", instanceID, err)))
			failed = append(failed, instanceID)
		} else {
			log.Info("Instance stopped successfully", "instance", instanceID)
			fmt.Println(ui.FormatOutput("✓ Success", fmt.Sprintf("Instance %s stopped successfully", instanceID)))
			successful = append(successful, instanceID)
		}
	}

	// Summary
	fmt.Println()
	if len(successful) > 0 {
		fmt.Printf("Successfully stopped %d instance(s): %s\n", 
			len(successful), strings.Join(successful, ", "))
	}
	if len(failed) > 0 {
		fmt.Printf("Failed to stop %d instance(s): %s\n", 
			len(failed), strings.Join(failed, ", "))
	}
}

func stopInstance(ctx context.Context, client *ec2.Client, instanceID string) error {
	// First, check if the instance exists and get its current state
	describeResult, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(describeResult.Reservations) == 0 || len(describeResult.Reservations[0].Instances) == 0 {
		return fmt.Errorf("instance not found")
	}

	instance := describeResult.Reservations[0].Instances[0]
	currentState := string(instance.State.Name)

	// Check if instance is already stopped or stopping
	if currentState == "stopped" {
		return fmt.Errorf("instance is already stopped")
	}
	if currentState == "stopping" {
		return fmt.Errorf("instance is already being stopped")
	}
	if currentState == "terminated" || currentState == "terminating" {
		return fmt.Errorf("instance is terminated or being terminated")
	}

	// Stop the instance
	_, err = client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return nil
}
