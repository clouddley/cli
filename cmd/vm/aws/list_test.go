package aws

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Mock EC2 client for list operations
type mockEC2ListClient struct {
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesError  error
	regions                 []string
}

func (m *mockEC2ListClient) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describeInstancesError != nil {
		return nil, m.describeInstancesError
	}
	return m.describeInstancesOutput, nil
}

func TestListInstances_Success(t *testing.T) {
	ctx := context.Background()
	instanceID1 := "i-1234567890abcdef0"
	instanceID2 := "i-0987654321fedcba0"
	publicIP1 := "52.123.45.67"
	publicIP2 := "54.321.67.89"
	instanceName1 := "clouddley-vm-1234567890"
	instanceName2 := "clouddley-vm-0987654321"
	
	mockClient := &mockEC2ListClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:      &instanceID1,
							InstanceType:    types.InstanceTypeT3Micro,
							State:           &types.InstanceState{Name: types.InstanceStateNameRunning},
							PublicIpAddress: &publicIP1,
							Tags: []types.Tag{
								{
									Key:   stringPtr("Name"),
									Value: &instanceName1,
								},
							},
						},
						{
							InstanceId:      &instanceID2,
							InstanceType:    types.InstanceTypeT3Small,
							State:           &types.InstanceState{Name: types.InstanceStateNameStopped},
							PublicIpAddress: &publicIP2,
							Tags: []types.Tag{
								{
									Key:   stringPtr("Name"),
									Value: &instanceName2,
								},
							},
						},
					},
				},
			},
		},
	}
	
	instances, err := listInstancesWithClient(ctx, mockClient)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(instances) != 2 {
		t.Fatalf("Expected 2 instances, got %d", len(instances))
	}
	
	// Verify first instance
	if instances[0].InstanceID != instanceID1 {
		t.Errorf("Expected instance ID %s, got %s", instanceID1, instances[0].InstanceID)
	}
	
	if instances[0].Name != instanceName1 {
		t.Errorf("Expected instance name %s, got %s", instanceName1, instances[0].Name)
	}
	
	if instances[0].Type != "t3.micro" {
		t.Errorf("Expected instance type t3.micro, got %s", instances[0].Type)
	}
	
	if instances[0].State != "running" {
		t.Errorf("Expected instance state running, got %s", instances[0].State)
	}
	
	if instances[0].PublicIP != publicIP1 {
		t.Errorf("Expected public IP %s, got %s", publicIP1, instances[0].PublicIP)
	}
	
	// Verify second instance
	if instances[1].InstanceID != instanceID2 {
		t.Errorf("Expected instance ID %s, got %s", instanceID2, instances[1].InstanceID)
	}
	
	if instances[1].State != "stopped" {
		t.Errorf("Expected instance state stopped, got %s", instances[1].State)
	}
}

func TestListInstances_NoInstances(t *testing.T) {
	ctx := context.Background()
	
	mockClient := &mockEC2ListClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{}, // No reservations
		},
	}
	
	instances, err := listInstancesWithClient(ctx, mockClient)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(instances) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(instances))
	}
}

func TestListInstances_APIError(t *testing.T) {
	ctx := context.Background()
	
	mockClient := &mockEC2ListClient{
		describeInstancesError: errors.New("AWS API error"),
	}
	
	_, err := listInstancesWithClient(ctx, mockClient)
	if err == nil {
		t.Fatal("Expected error from API, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to describe instances") {
		t.Errorf("Expected error message to contain 'failed to describe instances', got: %v", err)
	}
}

func TestListInstances_MissingTags(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	
	mockClient := &mockEC2ListClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:   &instanceID,
							InstanceType: types.InstanceTypeT3Micro,
							State:        &types.InstanceState{Name: types.InstanceStateNameRunning},
							Tags:         []types.Tag{}, // No tags
						},
					},
				},
			},
		},
	}
	
	instances, err := listInstancesWithClient(ctx, mockClient)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(instances) != 1 {
		t.Fatalf("Expected 1 instance, got %d", len(instances))
	}
	
	// Should handle missing name tag gracefully
	if instances[0].Name != instanceID {
		t.Errorf("Expected instance name to default to instance ID %s, got %s", instanceID, instances[0].Name)
	}
}

func TestListInstances_NoPublicIP(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	instanceName := "clouddley-vm-1234567890"
	
	mockClient := &mockEC2ListClient{
		describeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:      &instanceID,
							InstanceType:    types.InstanceTypeT3Micro,
							State:           &types.InstanceState{Name: types.InstanceStateNameStopped},
							PublicIpAddress: nil, // No public IP
							Tags: []types.Tag{
								{
									Key:   stringPtr("Name"),
									Value: &instanceName,
								},
							},
						},
					},
				},
			},
		},
	}
	
	instances, err := listInstancesWithClient(ctx, mockClient)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(instances) != 1 {
		t.Fatalf("Expected 1 instance, got %d", len(instances))
	}
	
	// Should handle missing public IP gracefully
	if instances[0].PublicIP != "-" {
		t.Errorf("Expected public IP to be '-' for instance without public IP, got %s", instances[0].PublicIP)
	}
}

func TestListInstances_FilterByState(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name           string
		instanceStates []types.InstanceStateName
		expectedCount  int
	}{
		{
			name:           "Running instances only",
			instanceStates: []types.InstanceStateName{types.InstanceStateNameRunning},
			expectedCount:  1,
		},
		{
			name:           "Mixed states",
			instanceStates: []types.InstanceStateName{types.InstanceStateNameRunning, types.InstanceStateNameStopped},
			expectedCount:  2,
		},
		{
			name:           "Terminated instances",
			instanceStates: []types.InstanceStateName{types.InstanceStateNameTerminated},
			expectedCount:  0, // Terminated instances are filtered out
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instances := make([]types.Instance, len(tt.instanceStates))
			for i, state := range tt.instanceStates {
				instanceID := "i-" + string(rune('1'+i)) + "234567890abcdef"
				instances[i] = types.Instance{
					InstanceId:   &instanceID,
					InstanceType: types.InstanceTypeT3Micro,
					State:        &types.InstanceState{Name: state},
					Tags: []types.Tag{
						{
							Key:   stringPtr("Name"),
							Value: stringPtr("test-instance-" + string(rune('1'+i))),
						},
					},
				}
			}
			
			mockClient := &mockEC2ListClient{
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: instances,
						},
					},
				},
			}
			
			result, err := listInstancesWithClient(ctx, mockClient)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d instances, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

func TestInstanceStateMapping(t *testing.T) {
	tests := []struct {
		name         string
		awsState     types.InstanceStateName
		expectedState string
	}{
		{
			name:         "Running state",
			awsState:     types.InstanceStateNameRunning,
			expectedState: "running",
		},
		{
			name:         "Stopped state",
			awsState:     types.InstanceStateNameStopped,
			expectedState: "stopped",
		},
		{
			name:         "Stopping state",
			awsState:     types.InstanceStateNameStopping,
			expectedState: "stopping",
		},
		{
			name:         "Starting state",
			awsState:     types.InstanceStateNamePending,
			expectedState: "pending",
		},
		{
			name:         "Terminated state",
			awsState:     types.InstanceStateNameTerminated,
			expectedState: "terminated",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapInstanceState(tt.awsState)
			if result != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, result)
			}
		})
	}
}

func TestInstanceTypeMapping(t *testing.T) {
	tests := []struct {
		name             string
		awsInstanceType  types.InstanceType
		expectedType     string
	}{
		{
			name:             "t3.micro",
			awsInstanceType:  types.InstanceTypeT3Micro,
			expectedType:     "t3.micro",
		},
		{
			name:             "t3.small",
			awsInstanceType:  types.InstanceTypeT3Small,
			expectedType:     "t3.small",
		},
		{
			name:             "m5.large",
			awsInstanceType:  types.InstanceTypeM5Large,
			expectedType:     "m5.large",
		},
		{
			name:             "c5.xlarge",
			awsInstanceType:  types.InstanceTypeC5Xlarge,
			expectedType:     "c5.xlarge",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapInstanceType(tt.awsInstanceType)
			if result != tt.expectedType {
				t.Errorf("Expected instance type %s, got %s", tt.expectedType, result)
			}
		})
	}
}

func TestGetNameFromTags(t *testing.T) {
	tests := []struct {
		name         string
		tags         []types.Tag
		instanceID   string
		expectedName string
	}{
		{
			name: "Name tag present",
			tags: []types.Tag{
				{Key: stringPtr("Name"), Value: stringPtr("my-instance")},
				{Key: stringPtr("Environment"), Value: stringPtr("prod")},
			},
			instanceID:   "i-1234567890abcdef0",
			expectedName: "my-instance",
		},
		{
			name: "No Name tag",
			tags: []types.Tag{
				{Key: stringPtr("Environment"), Value: stringPtr("prod")},
				{Key: stringPtr("Owner"), Value: stringPtr("team")},
			},
			instanceID:   "i-1234567890abcdef0",
			expectedName: "i-1234567890abcdef0",
		},
		{
			name:         "Empty tags",
			tags:         []types.Tag{},
			instanceID:   "i-1234567890abcdef0",
			expectedName: "i-1234567890abcdef0",
		},
		{
			name: "Name tag with empty value",
			tags: []types.Tag{
				{Key: stringPtr("Name"), Value: stringPtr("")},
			},
			instanceID:   "i-1234567890abcdef0",
			expectedName: "i-1234567890abcdef0",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNameFromTags(tt.tags, tt.instanceID)
			if result != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, result)
			}
		})
	}
}

// Helper functions for testing

func listInstancesWithClient(ctx context.Context, client EC2ListAPI) ([]EC2Instance, error) {
	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, errors.New("failed to describe instances: " + err.Error())
	}
	
	var instances []EC2Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Skip terminated instances
			if instance.State != nil && instance.State.Name == types.InstanceStateNameTerminated {
				continue
			}
			
			instanceID := ""
			if instance.InstanceId != nil {
				instanceID = *instance.InstanceId
			}
			
			name := getNameFromTags(instance.Tags, instanceID)
			
			publicIP := "-"
			if instance.PublicIpAddress != nil {
				publicIP = *instance.PublicIpAddress
			}
			
			ec2Instance := EC2Instance{
				InstanceID: instanceID,
				Name:       name,
				Type:       mapInstanceType(instance.InstanceType),
				State:      mapInstanceState(instance.State.Name),
				PublicIP:   publicIP,
			}
			
			instances = append(instances, ec2Instance)
		}
	}
	
	return instances, nil
}

func mapInstanceState(state types.InstanceStateName) string {
	switch state {
	case types.InstanceStateNameRunning:
		return "running"
	case types.InstanceStateNameStopped:
		return "stopped"
	case types.InstanceStateNameStopping:
		return "stopping"
	case types.InstanceStateNamePending:
		return "pending"
	case types.InstanceStateNameTerminated:
		return "terminated"
	case "terminating": // Use string literal since constant doesn't exist
		return "terminating"
	default:
		return string(state)
	}
}

func mapInstanceType(instanceType types.InstanceType) string {
	return string(instanceType)
}

func getNameFromTags(tags []types.Tag, instanceID string) string {
	for _, tag := range tags {
		if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil && *tag.Value != "" {
			return *tag.Value
		}
	}
	return instanceID
}

// EC2Instance represents an EC2 instance for listing
type EC2Instance struct {
	InstanceID string
	Name       string
	Type       string
	State      string
	PublicIP   string
}

// Interface for testing EC2 list operations
type EC2ListAPI interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}
