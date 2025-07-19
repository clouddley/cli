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

var deleteCmd = &cobra.Command{
	Use:     "delete --id <instance-id1,instance-id2,...> [--yes]",
	Aliases: []string{"del", "d"},
	Short:   "Delete (terminate) one or more AWS EC2 instances",
	Long:    `Delete (terminate) one or more AWS EC2 instances that were created by the Clouddley CLI. Supports comma-separated instance IDs.`,
	Run:     runDelete,
}

func init() {
	deleteCmd.Flags().StringP("id", "i", "", "Instance ID(s) to delete - supports comma-separated list (required)")
	deleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	deleteCmd.MarkFlagRequired("id")
}

func runDelete(cmd *cobra.Command, args []string) {
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
			fmt.Printf("Are you sure you want to terminate instance %s? (y/n): ", instanceIDs[0])
		} else {
			fmt.Printf("Are you sure you want to terminate %d instances (%s)? (y/n): ", 
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

	// Terminate instances
	for _, instanceID := range instanceIDs {
		fmt.Printf("Terminating instance %s...\n", instanceID)
		
		if err := terminateInstance(ctx, client, instanceID); err != nil {
			log.Error("Failed to terminate instance", "instance", instanceID, "error", err)
			fmt.Println(ui.FormatError(fmt.Sprintf("Failed to terminate %s: %v", instanceID, err)))
			failed = append(failed, instanceID)
		} else {
			log.Info("Instance terminated successfully", "instance", instanceID)
			fmt.Println(ui.FormatOutput("✓ Success", fmt.Sprintf("Instance %s terminated successfully", instanceID)))
			successful = append(successful, instanceID)
		}
	}

	// Summary
	fmt.Println()
	if len(successful) > 0 {
		fmt.Printf("Successfully terminated %d instance(s): %s\n", 
			len(successful), strings.Join(successful, ", "))
	}
	if len(failed) > 0 {
		fmt.Printf("Failed to terminate %d instance(s): %s\n", 
			len(failed), strings.Join(failed, ", "))
	}
}

func terminateInstance(ctx context.Context, client *ec2.Client, instanceID string) error {
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

	// Check if instance is already terminated or terminating
	if currentState == "terminated" {
		return fmt.Errorf("instance is already terminated")
	}
	if currentState == "terminating" {
		return fmt.Errorf("instance is already being terminated")
	}

	// Terminate the instance
	_, err = client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	return nil
}
