package aws

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Mock EC2 client for stop operations
type mockEC2StopClient struct {
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesError  error
	stopInstancesOutput     *ec2.StopInstancesOutput
	stopInstancesError      error
	stoppedInstances        []string
}

func (m *mockEC2StopClient) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describeInstancesError != nil {
		return nil, m.describeInstancesError
	}
	return m.describeInstancesOutput, nil
}

func (m *mockEC2StopClient) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	if m.stopInstancesError != nil {
		return nil, m.stopInstancesError
	}
	
	// Record stopped instance
	if len(params.InstanceIds) > 0 {
		m.stoppedInstances = append(m.stoppedInstances, params.InstanceIds...)
	}
	
	return m.stopInstancesOutput, nil
}

func TestStopInstance_Success(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StopClient{
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
		stopInstancesOutput: &ec2.StopInstancesOutput{},
	}
	
	err := stopInstanceWithClient(ctx, mockClient, instanceID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(mockClient.stoppedInstances) != 1 {
		t.Fatalf("Expected 1 stopped instance, got %d", len(mockClient.stoppedInstances))
	}
	
	if mockClient.stoppedInstances[0] != instanceID {
		t.Errorf("Expected stopped instance %s, got %s", instanceID, mockClient.stoppedInstances[0])
	}
}

func TestStopInstance_AlreadyStopped(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StopClient{
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
	}
	
	err := stopInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for already stopped instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "already stopped") {
		t.Errorf("Expected error message to contain 'already stopped', got: %v", err)
	}
	
	if len(mockClient.stoppedInstances) != 0 {
		t.Errorf("Expected no instances to be stopped, got %d", len(mockClient.stoppedInstances))
	}
}

func TestStopInstance_AlreadyStopping(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StopClient{
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
	}
	
	err := stopInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for already stopping instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "already being stopped") {
		t.Errorf("Expected error message to contain 'already being stopped', got: %v", err)
	}
}

func TestStopInstance_Terminated(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StopClient{
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
	
	err := stopInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for terminated instance, got nil")
	}
	
	if !strings.Contains(err.Error(), "terminated") {
		t.Errorf("Expected error message to contain 'terminated', got: %v", err)
	}
}

func TestStopInstance_NotFound(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StopClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{}, // No reservations
		},
	}
	
	err := stopInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error for instance not found, got nil")
	}
	
	if !strings.Contains(err.Error(), "instance not found") {
		t.Errorf("Expected error message to contain 'instance not found', got: %v", err)
	}
}

func TestStopInstance_DescribeError(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StopClient{
		describeInstancesError: errors.New("AWS API error"),
	}
	
	err := stopInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error from DescribeInstances, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to describe instance") {
		t.Errorf("Expected error message to contain 'failed to describe instance', got: %v", err)
	}
}

func TestStopInstance_StopError(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2StopClient{
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
		stopInstancesError: errors.New("stop API error"),
	}
	
	err := stopInstanceWithClient(ctx, mockClient, instanceID)
	if err == nil {
		t.Fatal("Expected error from StopInstances, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to stop instance") {
		t.Errorf("Expected error message to contain 'failed to stop instance', got: %v", err)
	}
}

func TestStopInstance_StateValidation(t *testing.T) {
	tests := []struct {
		name        string
		state       types.InstanceStateName
		expectError bool
		errorText   string
	}{
		{
			name:        "Running instance",
			state:       types.InstanceStateNameRunning,
			expectError: false,
		},
		{
			name:        "Pending instance",
			state:       types.InstanceStateNamePending,
			expectError: false,
		},
		{
			name:        "Stopped instance",
			state:       types.InstanceStateNameStopped,
			expectError: true,
			errorText:   "already stopped",
		},
		{
			name:        "Stopping instance",
			state:       types.InstanceStateNameStopping,
			expectError: true,
			errorText:   "already being stopped",
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
			
			mockClient := &mockEC2StopClient{
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
				stopInstancesOutput: &ec2.StopInstancesOutput{},
			}
			
			err := stopInstanceWithClient(ctx, mockClient, instanceID)
			
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

func TestStopInstance_MultipleInstances(t *testing.T) {
	ctx := context.Background()
	instanceIDs := []string{"i-1234567890abcdef0", "i-0987654321fedcba0", "i-1111222233334444"}
	
	// Test stopping multiple instances in different scenarios
	tests := []struct {
		name                string
		instanceID          string
		state               types.InstanceStateName
		stopError           error
		expectStopCall      bool
		expectError         bool
	}{
		{
			name:           "First instance running",
			instanceID:     instanceIDs[0],
			state:          types.InstanceStateNameRunning,
			expectStopCall: true,
			expectError:    false,
		},
		{
			name:           "Second instance already stopped",
			instanceID:     instanceIDs[1],
			state:          types.InstanceStateNameStopped,
			expectStopCall: false,
			expectError:    true,
		},
		{
			name:           "Third instance running with stop error",
			instanceID:     instanceIDs[2],
			state:          types.InstanceStateNameRunning,
			stopError:      errors.New("stop failed"),
			expectStopCall: true,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockEC2StopClient{
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
				stopInstancesOutput: &ec2.StopInstancesOutput{},
				stopInstancesError:  tt.stopError,
			}
			
			err := stopInstanceWithClient(ctx, mockClient, tt.instanceID)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for instance %s, got nil", tt.instanceID)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for instance %s, got: %v", tt.instanceID, err)
				}
			}
			
			if tt.expectStopCall {
				if len(mockClient.stoppedInstances) == 0 && tt.stopError == nil {
					t.Errorf("Expected stop call for instance %s, but no instances were stopped", tt.instanceID)
				}
			} else {
				if len(mockClient.stoppedInstances) > 0 {
					t.Errorf("Did not expect stop call for instance %s, but instances were stopped", tt.instanceID)
				}
			}
		})
	}
}

// Helper function for testing
func stopInstanceWithClient(ctx context.Context, client EC2StopAPI, instanceID string) error {
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

	// Check if instance is already stopped or stopping
	if currentState == "stopped" {
		return errors.New("instance is already stopped")
	}
	if currentState == "stopping" {
		return errors.New("instance is already being stopped")
	}
	if currentState == "terminated" || currentState == "terminating" {
		return errors.New("instance is terminated or being terminated")
	}

	// Stop the instance
	_, err = client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return errors.New("failed to stop instance: " + err.Error())
	}

	return nil
}

// Interface for testing EC2 stop operations
type EC2StopAPI interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
}
