package aws

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Mock EC2 client for start operations
type mockEC2StartClient struct {
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesError  error
	startInstancesOutput    *ec2.StartInstancesOutput
	startInstancesError     error
	startedInstances        []string
}

func (m *mockEC2StartClient) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describeInstancesError != nil {
		return nil, m.describeInstancesError
	}
	return m.describeInstancesOutput, nil
}

func (m *mockEC2StartClient) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	if m.startInstancesError != nil {
		return nil, m.startInstancesError
	}
	
	// Record started instance
	if len(params.InstanceIds) > 0 {
		m.startedInstances = append(m.startedInstances, params.InstanceIds...)
	}
	
	return m.startInstancesOutput, nil
}

func TestStartInstance_Success(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State:      &types.InstanceState{Name: types.InstanceStateNameStopped},
						},
					},
				},
			},
		},
		startInstancesOutput: &ec2.StartInstancesOutput{},
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(mockClient.startedInstances) != 1 {
		t.Fatalf("Expected 1 started instance, got %d", len(mockClient.startedInstances))
	}
	
	if mockClient.startedInstances[0] != instanceID {
		t.Errorf("Expected started instance %s, got %s", instanceID, mockClient.startedInstances[0])
	}
}

func TestStartInstance_AlreadyRunning(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State:      &types.InstanceState{Name: types.InstanceStateNameRunning},
						},
					},
				},
			},
		},
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for already running instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("Expected error message to contain 'already running', got: %v", err)
	}
	
	if len(mockClient.startedInstances) != 0 {
		t.Errorf("Expected no instances to be started, got %d", len(mockClient.startedInstances))
	}
}

func TestStartInstance_AlreadyStarting(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State:      &types.InstanceState{Name: types.InstanceStateNamePending},
						},
					},
				},
			},
		},
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for already starting instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "already starting") {
		t.Errorf("Expected error message to contain 'already starting', got: %v", err)
	}
}

func TestStartInstance_Terminated(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State:      &types.InstanceState{Name: types.InstanceStateNameTerminated},
						},
					},
				},
			},
		},
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for terminated instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "terminated") {
		t.Errorf("Expected error message to contain 'terminated', got: %v", err)
	}
}

func TestStartInstance_NotFound(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{}, // No reservations
		},
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for instance not found, got nil")
	}
	
	if !strings.Contains(err.Error(), "instance not found") {
		t.Errorf("Expected error message to contain 'instance not found', got: %v", err)
	}
}

func TestStartInstance_DescribeError(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesError: errors.New("AWS API error"),
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error from DescribeInstances, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to describe instance") {
		t.Errorf("Expected error message to contain 'failed to describe instance', got: %v", err)
	}
}

func TestStartInstance_StartError(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State:      &types.InstanceState{Name: types.InstanceStateNameStopped},
						},
					},
				},
			},
		},
		startInstancesError: errors.New("start API error"),
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error from StartInstances, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to start instance") {
		t.Errorf("Expected error message to contain 'failed to start instance', got: %v", err)
	}
}

func TestStartInstance_StateValidation(t *testing.T) {
	tests := []struct {
		name        string
		state       types.InstanceStateName
		expectError bool
		errorText   string
	}{
		{
			name:        "Stopped instance",
			state:       types.InstanceStateNameStopped,
			expectError: false,
		},
		{
			name:        "Stopping instance",
			state:       types.InstanceStateNameStopping,
			expectError: false,
		},
		{
			name:        "Running instance",
			state:       types.InstanceStateNameRunning,
			expectError: true,
			errorText:   "already running",
		},
		{
			name:        "Pending instance",
			state:       types.InstanceStateNamePending,
			expectError: true,
			errorText:   "already starting",
		},
		{
			name:        "Terminated instance",
			state:       types.InstanceStateNameTerminated,
			expectError: true,
			errorText:   "terminated",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			instanceID := "i-1234567890abcdef0"
			
			mockClient := &mockEC2StartClient{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: &instanceID,
									State:      &types.InstanceState{Name: tt.state},
								},
							},
						},
					},
				},
				startInstancesOutput: &ec2.StartInstancesOutput{},
			}
			
			err := startInstanceWithClient(ctx, mockClient, instanceID)
			
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error for state %s, got nil", tt.state)
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for state %s, got: %v", tt.state, err)
				}
			}
		})
	}
}

func TestStartInstance_MultipleInstances(t *testing.T) {
	ctx := context.Background()
	instanceIDs := []string{"i-1234567890abcdef0", "i-0987654321fedcba0", "i-1111222233334444"}
	
	// Test starting multiple instances in different scenarios
	tests := []struct {
		name                string
		instanceID          string
		state               types.InstanceStateName
		startError          error
		expectStartCall     bool
		expectError         bool
	}{
		{
			name:            "First instance stopped",
			instanceID:      instanceIDs[0],
			state:           types.InstanceStateNameStopped,
			expectStartCall: true,
			expectError:     false,
		},
		{
			name:            "Second instance already running",
			instanceID:      instanceIDs[1],
			state:           types.InstanceStateNameRunning,
			expectStartCall: false,
			expectError:     true,
		},
		{
			name:            "Third instance stopped with start error",
			instanceID:      instanceIDs[2],
			state:           types.InstanceStateNameStopped,
			startError:      errors.New("start failed"),
			expectStartCall: true,
			expectError:     true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockEC2StartClient{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: &tt.instanceID,
									State:      &types.InstanceState{Name: tt.state},
								},
							},
						},
					},
				},
				startInstancesOutput: &ec2.StartInstancesOutput{},
				startInstancesError:  tt.startError,
			}
			
			err := startInstanceWithClient(ctx, mockClient, tt.instanceID)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for instance %s, got nil", tt.instanceID)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for instance %s, got: %v", tt.instanceID, err)
				}
			}
			
			if tt.expectStartCall {
				if len(mockClient.startedInstances) == 0 && tt.startError == nil {
					t.Errorf("Expected start call for instance %s, but no instances were started", tt.instanceID)
				}
			} else {
				if len(mockClient.startedInstances) > 0 {
					t.Errorf("Did not expect start call for instance %s, but instances were started", tt.instanceID)
				}
			}
		})
	}
}

func TestStartInstance_FromStoppingState(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State:      &types.InstanceState{Name: types.InstanceStateNameStopping},
						},
					},
				},
			},
		},
		startInstancesOutput: &ec2.StartInstancesOutput{},
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err != nil {
		t.Fatalf("Expected no error for stopping instance, got: %v", err)
	}
	
	if len(mockClient.startedInstances) != 1 {
		t.Fatalf("Expected 1 started instance, got %d", len(mockClient.startedInstances))
	}
	
	if mockClient.startedInstances[0] != instanceID {
		t.Errorf("Expected started instance %s, got %s", instanceID, mockClient.startedInstances[0])
	}
}

func TestStartInstance_TerminatingState(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	// Create a custom state name for terminating since it might not be a predefined constant
	terminatingState := types.InstanceStateName("terminating")
	
	mockClient := &mockEC2StartClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: &instanceID,
							State:      &types.InstanceState{Name: terminatingState},
						},
					},
				},
			},
		},
	}
	
	err := startInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for terminating instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "terminated") {
		t.Errorf("Expected error message to contain 'terminated', got: %v", err)
	}
}

// Helper function for testing
func startInstanceWithClient(ctx context.Context, client EC2StartAPI, instanceID string) error {
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

	// Check if instance is already running or pending
	if currentState == "running" {
		return errors.New("instance is already running")
	}
	if currentState == "pending" {
		return errors.New("instance is already starting")
	}
	if currentState == "terminated" || currentState == "terminating" {
		return errors.New("instance is terminated or being terminated")
	}

	// Start the instance
	_, err = client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return errors.New("failed to start instance: " + err.Error())
	}

	return nil
}

// Interface for testing EC2 start operations
type EC2StartAPI interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
}
