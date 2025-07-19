package aws

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Mock EC2 client for testing
type mockEC2Client struct {
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesError  error
	terminateInstancesError error
	terminatedInstances     []string
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describeInstancesError != nil {
		return nil, m.describeInstancesError
	}
	return m.describeInstancesOutput, nil
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if m.terminateInstancesError != nil {
		return nil, m.terminateInstancesError
	}
	
	// Record which instances were requested for termination
	if len(params.InstanceIds) > 0 {
		m.terminatedInstances = append(m.terminatedInstances, params.InstanceIds[0])
	}
	
	return &ec2.TerminateInstancesOutput{}, nil
}

func TestTerminateInstance_Success(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2Client{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				},
			},
		},
	}
	
	err := terminateInstanceWithClient(ctx, mockClient, instanceID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify that terminate was called
	if len(mockClient.terminatedInstances) != 1 {
		t.Fatalf("Expected 1 instance to be terminated, got %d", len(mockClient.terminatedInstances))
	}
	
	if mockClient.terminatedInstances[0] != instanceID {
		t.Errorf("Expected instance %s to be terminated, got %s", instanceID, mockClient.terminatedInstances[0])
	}
}

func TestTerminateInstance_InstanceNotFound(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-nonexistent"
	
	mockClient := &mockEC2Client{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{}, // Empty - instance not found
		},
	}
	
	err := terminateInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for non-existent instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "instance not found") {
		t.Errorf("Expected error message to contain 'instance not found', got: %v", err)
	}
}

func TestTerminateInstance_AlreadyTerminated(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2Client{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State: &types.InstanceState{
								Name: types.InstanceStateNameTerminated,
							},
						},
					},
				},
			},
		},
	}
	
	err := terminateInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for already terminated instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "already terminated") {
		t.Errorf("Expected error message to contain 'already terminated', got: %v", err)
	}
}

func TestTerminateInstance_AlreadyTerminating(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2Client{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State: &types.InstanceState{
								Name: "terminating", // Use string literal since constant doesn't exist
							},
						},
					},
				},
			},
		},
	}
	
	err := terminateInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for already terminating instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "already being terminated") {
		t.Errorf("Expected error message to contain 'already being terminated', got: %v", err)
	}
}

func TestTerminateInstance_DescribeError(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2Client{
		describeInstancesError: errors.New("AWS API error"),
	}
	
	err := terminateInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error from describe instances, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to describe instance") {
		t.Errorf("Expected error message to contain 'failed to describe instance', got: %v", err)
	}
}

func TestTerminateInstance_TerminateError(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2Client{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				},
			},
		},
		terminateInstancesError: errors.New("termination failed"),
	}
	
	err := terminateInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error from terminate instances, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to terminate instance") {
		t.Errorf("Expected error message to contain 'failed to terminate instance', got: %v", err)
	}
}

func TestParseInstanceIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single instance ID",
			input:    "i-1234567890abcdef0",
			expected: []string{"i-1234567890abcdef0"},
		},
		{
			name:     "Multiple instance IDs",
			input:    "i-1234567890abcdef0,i-0987654321fedcba0",
			expected: []string{"i-1234567890abcdef0", "i-0987654321fedcba0"},
		},
		{
			name:     "Instance IDs with spaces",
			input:    "i-1234567890abcdef0, i-0987654321fedcba0",
			expected: []string{"i-1234567890abcdef0", "i-0987654321fedcba0"},
		},
		{
			name:     "Instance IDs with extra spaces",
			input:    "i-1234567890abcdef0,   i-0987654321fedcba0",
			expected: []string{"i-1234567890abcdef0", "i-0987654321fedcba0"},
		},
		{
			name:     "Instance IDs with trailing comma and space",
			input:    "i-1234567890abcdef0, ",
			expected: []string{"i-1234567890abcdef0"},
		},
		{
			name:     "Instance IDs with multiple trailing commas",
			input:    "i-1234567890abcdef0,,,",
			expected: []string{"i-1234567890abcdef0"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Only commas",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "Only spaces and commas",
			input:    " , , , ",
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInstanceIDs(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d instance IDs, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected instance ID %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

func TestValidateInstanceIDs(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		expectErr bool
	}{
		{
			name:      "Valid instance IDs",
			input:     []string{"i-1234567890abcdef0", "i-0987654321fedcba0"},
			expectErr: false,
		},
		{
			name:      "Single valid instance ID",
			input:     []string{"i-1234567890abcdef0"},
			expectErr: false,
		},
		{
			name:      "Empty slice",
			input:     []string{},
			expectErr: true,
		},
		{
			name:      "Empty string in slice",
			input:     []string{"i-1234567890abcdef0", ""},
			expectErr: true,
		},
		{
			name:      "Invalid instance ID format",
			input:     []string{"invalid-id"},
			expectErr: true,
		},
		{
			name:      "Mixed valid and invalid",
			input:     []string{"i-1234567890abcdef0", "invalid-id"},
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInstanceIDs(tt.input)
			
			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// Helper functions for testing (these would be used by the actual implementation)

func terminateInstanceWithClient(ctx context.Context, client EC2TerminateAPI, instanceID string) error {
	// First, check if the instance exists and get its current state
	describeResult, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return errors.New("failed to describe instance: " + err.Error())
	}

	if len(describeResult.Reservations) == 0 || len(describeResult.Reservations[0].Instances) == 0 {
		return errors.New("instance not found")
	}

	instance := describeResult.Reservations[0].Instances[0]
	currentState := string(instance.State.Name)

	// Check if instance is already terminated or terminating
	if currentState == "terminated" {
		return errors.New("instance is already terminated")
	}
	if currentState == "terminating" {
		return errors.New("instance is already being terminated")
	}

	// Terminate the instance
	_, err = client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return errors.New("failed to terminate instance: " + err.Error())
	}

	return nil
}

func parseInstanceIDs(input string) []string {
	parts := strings.Split(input, ",")
	var instanceIDs []string
	for _, id := range parts {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" { // Filter out empty strings
			instanceIDs = append(instanceIDs, trimmed)
		}
	}
	return instanceIDs
}

func validateInstanceIDs(instanceIDs []string) error {
	if len(instanceIDs) == 0 {
		return errors.New("no instance IDs provided")
	}
	
	for _, id := range instanceIDs {
		if id == "" {
			return errors.New("empty instance ID found")
		}
		
		// Basic validation for AWS instance ID format
		if !strings.HasPrefix(id, "i-") || len(id) != 19 {
			return errors.New("invalid instance ID format: " + id)
		}
	}
	
	return nil
}

// Interface for testing EC2 terminate operations
type EC2TerminateAPI interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
}
