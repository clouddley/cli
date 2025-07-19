package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	awsinternal "github.com/clouddley/clouddley/internal/aws"
	"github.com/clouddley/clouddley/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS EC2 instances created by Clouddley CLI",
	Long:  `List all AWS EC2 instances that were created by the Clouddley CLI.`,
	Run:   runList,
}

func runList(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Validate AWS credentials
	if err := awsinternal.ValidateAWSCredentials(ctx); err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error: %v", err)))
		return
	}

	// Get AWS config for region info
	cfg, err := awsinternal.GetAWSConfig(ctx)
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error getting AWS config: %v", err)))
		return
	}

	// List instances
	instances, err := listCloudleyInstances(ctx)
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error listing instances: %v", err)))
		return
	}

	if len(instances) == 0 {
		fmt.Println(ui.FormatOutput("✓ Clouddley CLI Instances", fmt.Sprintf("No instances found in region %s", cfg.Region)))
		return
	}

	// Create and display table
	fmt.Printf("Clouddley CLI Instances in region %s:\n\n", cfg.Region)
	displayInstancesTable(instances)
}

type ClouddleyInstance struct {
	InstanceID   string
	Name         string
	State        string
	InstanceType string
	PublicIP     string
	LaunchTime   string
}

func listCloudleyInstances(ctx context.Context) ([]ClouddleyInstance, error) {
	client, err := awsinternal.GetEC2Client(ctx)
	if err != nil {
		return nil, err
	}

	// Describe instances with Clouddley tag
	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:CreatedBy"),
				Values: []string{"Clouddley"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instances []ClouddleyInstance
	
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Skip terminated instances
			if instance.State.Name == types.InstanceStateNameTerminated {
				continue
			}

			var name string
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
					break
				}
			}

			publicIP := "N/A"
			if instance.PublicIpAddress != nil {
				publicIP = *instance.PublicIpAddress
			}

			launchTime := "N/A"
			if instance.LaunchTime != nil {
				launchTime = instance.LaunchTime.Format("2006-01-02 15:04:05")
			}

			instances = append(instances, ClouddleyInstance{
				InstanceID:   *instance.InstanceId,
				Name:         name,
				State:        string(instance.State.Name),
				InstanceType: string(instance.InstanceType),
				PublicIP:     publicIP,
				LaunchTime:   launchTime,
			})
		}
	}

	return instances, nil
}

func displayInstancesTable(instances []ClouddleyInstance) {
	columns := []table.Column{
		{Title: "Instance ID", Width: 20},
		{Title: "Name", Width: 20},
		{Title: "State", Width: 12},
		{Title: "Type", Width: 12},
		{Title: "Public IP", Width: 15},
		{Title: "Launch Time", Width: 20},
	}

	rows := make([]table.Row, len(instances))
	for i, instance := range instances {
		rows[i] = table.Row{
			instance.InstanceID,
			instance.Name,
			instance.State,
			instance.InstanceType,
			instance.PublicIP,
			instance.LaunchTime,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(len(instances)+2),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))
	
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("#FAFAFA"))

	t.SetStyles(s)

	fmt.Println(t.View())
}
