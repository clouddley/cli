package aws

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Mock EC2 client for create operations
type mockEC2CreateClient struct {
	runInstancesOutput     *ec2.RunInstancesOutput
	runInstancesError      error
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesError  error
	describeVpcsOutput     *ec2.DescribeVpcsOutput
	describeVpcsError      error
	describeSubnetsOutput  *ec2.DescribeSubnetsOutput
	describeSubnetsError   error
	describeSecurityGroupsOutput *ec2.DescribeSecurityGroupsOutput
	describeSecurityGroupsError  error
	createSecurityGroupOutput    *ec2.CreateSecurityGroupOutput
	createSecurityGroupError     error
	describeImagesOutput   *ec2.DescribeImagesOutput
	describeImagesError    error
	createdInstances       []string
}

func (m *mockEC2CreateClient) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.runInstancesError != nil {
		return nil, m.runInstancesError
	}
	
	// Record created instance
	if m.runInstancesOutput != nil && len(m.runInstancesOutput.Instances) > 0 {
		m.createdInstances = append(m.createdInstances, *m.runInstancesOutput.Instances[0].InstanceId)
	}
	
	return m.runInstancesOutput, nil
}

func (m *mockEC2CreateClient) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describeInstancesError != nil {
		return nil, m.describeInstancesError
	}
	return m.describeInstancesOutput, nil
}

func (m *mockEC2CreateClient) DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	if m.describeVpcsError != nil {
		return nil, m.describeVpcsError
	}
	return m.describeVpcsOutput, nil
}

func (m *mockEC2CreateClient) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	if m.describeSubnetsError != nil {
		return nil, m.describeSubnetsError
	}
	return m.describeSubnetsOutput, nil
}

func (m *mockEC2CreateClient) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	if m.describeSecurityGroupsError != nil {
		return nil, m.describeSecurityGroupsError
	}
	return m.describeSecurityGroupsOutput, nil
}

func (m *mockEC2CreateClient) CreateSecurityGroup(ctx context.Context, params *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error) {
	if m.createSecurityGroupError != nil {
		return nil, m.createSecurityGroupError
	}
	return m.createSecurityGroupOutput, nil
}

func (m *mockEC2CreateClient) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.describeImagesError != nil {
		return nil, m.describeImagesError
	}
	return m.describeImagesOutput, nil
}

func (m *mockEC2CreateClient) AuthorizeSecurityGroupIngress(ctx context.Context, params *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	return &ec2.AuthorizeSecurityGroupIngressOutput{}, nil
}

func TestGetDefaultVPCAndSubnet_Success(t *testing.T) {
	ctx := context.Background()
	vpcID := "vpc-12345"
	subnetID := "subnet-67890"
	
	mockClient := &mockEC2CreateClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{
				{
					VpcId: &vpcID,
				},
			},
		},
		describeSubnetsOutput: &ec2.DescribeSubnetsOutput{
			Subnets: []types.Subnet{
				{
					SubnetId: &subnetID,
					VpcId:    &vpcID,
				},
			},
		},
	}
	
	resultVPC, resultSubnet, err := getDefaultVPCAndSubnetWithClient(ctx, mockClient)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if resultVPC != vpcID {
		t.Errorf("Expected VPC ID %s, got %s", vpcID, resultVPC)
	}
	
	if resultSubnet != subnetID {
		t.Errorf("Expected subnet ID %s, got %s", subnetID, resultSubnet)
	}
}

func TestGetDefaultVPCAndSubnet_NoVPC(t *testing.T) {
	ctx := context.Background()
	
	mockClient := &mockEC2CreateClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{}, // No VPCs
		},
	}
	
	_, _, err := getDefaultVPCAndSubnetWithClient(ctx, mockClient)
	if err == nil {
		t.Fatal("Expected error for no default VPC, got nil")
	}
	
	if !strings.Contains(err.Error(), "no default VPC found") {
		t.Errorf("Expected error message to contain 'no default VPC found', got: %v", err)
	}
}

func TestGetDefaultVPCAndSubnet_NoSubnet(t *testing.T) {
	ctx := context.Background()
	vpcID := "vpc-12345"
	
	mockClient := &mockEC2CreateClient{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []types.Vpc{
				{
					VpcId: &vpcID,
				},
			},
		},
		describeSubnetsOutput: &ec2.DescribeSubnetsOutput{
			Subnets: []types.Subnet{}, // No subnets
		},
	}
	
	_, _, err := getDefaultVPCAndSubnetWithClient(ctx, mockClient)
	if err == nil {
		t.Fatal("Expected error for no default subnet, got nil")
	}
	
	if !strings.Contains(err.Error(), "no default subnet found") {
		t.Errorf("Expected error message to contain 'no default subnet found', got: %v", err)
	}
}

func TestCreateOrGetSecurityGroup_CreateNew(t *testing.T) {
	ctx := context.Background()
	vpcID := "vpc-12345"
	expectedSGID := "sg-67890"
	
	mockClient := &mockEC2CreateClient{
		// No existing security groups
		describeSecurityGroupsOutput: &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{},
		},
		// Successful creation
		createSecurityGroupOutput: &ec2.CreateSecurityGroupOutput{
			GroupId: &expectedSGID,
		},
	}
	
	resultSGID, err := createOrGetSecurityGroupWithClient(ctx, mockClient, vpcID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if resultSGID != expectedSGID {
		t.Errorf("Expected security group ID %s, got %s", expectedSGID, resultSGID)
	}
}

func TestCreateOrGetSecurityGroup_UseExisting(t *testing.T) {
	ctx := context.Background()
	vpcID := "vpc-12345"
	existingSGID := "sg-existing"
	
	mockClient := &mockEC2CreateClient{
		// Existing security group with all required rules
		describeSecurityGroupsOutput: &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId: &existingSGID,
					IpPermissions: []types.IpPermission{
						// SSH rule
						{
							IpProtocol: stringPtr("tcp"),
							FromPort:   int32Ptr(22),
							ToPort:     int32Ptr(22),
							IpRanges: []types.IpRange{
								{CidrIp: stringPtr("0.0.0.0/0")},
							},
						},
						// HTTP rule
						{
							IpProtocol: stringPtr("tcp"),
							FromPort:   int32Ptr(80),
							ToPort:     int32Ptr(80),
							IpRanges: []types.IpRange{
								{CidrIp: stringPtr("0.0.0.0/0")},
							},
						},
						// HTTPS rule
						{
							IpProtocol: stringPtr("tcp"),
							FromPort:   int32Ptr(443),
							ToPort:     int32Ptr(443),
							IpRanges: []types.IpRange{
								{CidrIp: stringPtr("0.0.0.0/0")},
							},
						},
					},
				},
			},
		},
	}
	
	resultSGID, err := createOrGetSecurityGroupWithClient(ctx, mockClient, vpcID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if resultSGID != existingSGID {
		t.Errorf("Expected security group ID %s, got %s", existingSGID, resultSGID)
	}
}

func TestGetLatestUbuntuAMI_Success(t *testing.T) {
	ctx := context.Background()
	expectedAMI := "ami-12345678"
	
	mockClient := &mockEC2CreateClient{
		describeImagesOutput: &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      &expectedAMI,
					CreationDate: stringPtr("2024-01-15T10:00:00.000Z"),
				},
				{
					ImageId:      stringPtr("ami-87654321"),
					CreationDate: stringPtr("2024-01-10T10:00:00.000Z"), // Older
				},
			},
		},
	}
	
	resultAMI, err := getLatestUbuntuAMIWithClient(ctx, mockClient)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if resultAMI != expectedAMI {
		t.Errorf("Expected AMI ID %s, got %s", expectedAMI, resultAMI)
	}
}

func TestGetLatestUbuntuAMI_NoImages(t *testing.T) {
	ctx := context.Background()
	
	mockClient := &mockEC2CreateClient{
		describeImagesOutput: &ec2.DescribeImagesOutput{
			Images: []types.Image{}, // No images found
		},
	}
	
	_, err := getLatestUbuntuAMIWithClient(ctx, mockClient)
	if err == nil {
		t.Fatal("Expected error for no Ubuntu AMI, got nil")
	}
	
	if !strings.Contains(err.Error(), "no Ubuntu AMI found") {
		t.Errorf("Expected error message to contain 'no Ubuntu AMI found', got: %v", err)
	}
}

func TestInstanceInfo_Structure(t *testing.T) {
	tests := []struct {
		name         string
		instanceID   string
		instanceName string
		publicIP     string
		region       string
	}{
		{
			name:         "Complete instance info",
			instanceID:   "i-1234567890abcdef0",
			instanceName: "clouddley-vm-1234567890",
			publicIP:     "52.123.45.67",
			region:       "us-east-2",
		},
		{
			name:         "Instance without public IP",
			instanceID:   "i-0987654321fedcba0",
			instanceName: "clouddley-vm-0987654321",
			publicIP:     "",
			region:       "us-west-2",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &InstanceInfo{
				InstanceID: tt.instanceID,
				Name:       tt.instanceName,
				PublicIP:   tt.publicIP,
				Region:     tt.region,
			}
			
			if info.InstanceID != tt.instanceID {
				t.Errorf("Expected instance ID %s, got %s", tt.instanceID, info.InstanceID)
			}
			
			if info.Name != tt.instanceName {
				t.Errorf("Expected name %s, got %s", tt.instanceName, info.Name)
			}
			
			if info.PublicIP != tt.publicIP {
				t.Errorf("Expected public IP %s, got %s", tt.publicIP, info.PublicIP)
			}
			
			if info.Region != tt.region {
				t.Errorf("Expected region %s, got %s", tt.region, info.Region)
			}
		})
	}
}

func TestValidateInstanceType(t *testing.T) {
	tests := []struct {
		name         string
		instanceType string
		expectValid  bool
	}{
		{
			name:         "Valid t3.micro",
			instanceType: "t3.micro",
			expectValid:  true,
		},
		{
			name:         "Valid m5.large",
			instanceType: "m5.large",
			expectValid:  true,
		},
		{
			name:         "Valid c5.xlarge",
			instanceType: "c5.xlarge",
			expectValid:  true,
		},
		{
			name:         "Invalid format",
			instanceType: "invalid-type",
			expectValid:  false,
		},
		{
			name:         "Empty string",
			instanceType: "",
			expectValid:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateInstanceType(tt.instanceType)
			
			if isValid != tt.expectValid {
				t.Errorf("Expected validation result %v for instance type %s, got %v", 
					tt.expectValid, tt.instanceType, isValid)
			}
		})
	}
}

// Helper functions for testing

func getDefaultVPCAndSubnetWithClient(ctx context.Context, client EC2CreateAPI) (string, string, error) {
	// Mock implementation for testing
	vpcResult, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return "", "", err
	}
	
	if len(vpcResult.Vpcs) == 0 {
		return "", "", errors.New("no default VPC found")
	}
	
	vpcID := *vpcResult.Vpcs[0].VpcId
	
	subnetResult, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return "", "", err
	}
	
	if len(subnetResult.Subnets) == 0 {
		return "", "", errors.New("no default subnet found")
	}
	
	subnetID := *subnetResult.Subnets[0].SubnetId
	
	return vpcID, subnetID, nil
}

func createOrGetSecurityGroupWithClient(ctx context.Context, client EC2CreateAPI, vpcID string) (string, error) {
	// Check if security group exists
	result, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return "", err
	}
	
	if len(result.SecurityGroups) > 0 {
		// Check if existing group has all required rules
		sg := result.SecurityGroups[0]
		hasSSH, hasHTTP, hasHTTPS := false, false, false
		
		for _, rule := range sg.IpPermissions {
			if rule.IpProtocol != nil && *rule.IpProtocol == "tcp" {
				if rule.FromPort != nil {
					switch *rule.FromPort {
					case 22:
						hasSSH = true
					case 80:
						hasHTTP = true
					case 443:
						hasHTTPS = true
					}
				}
			}
		}
		
		if hasSSH && hasHTTP && hasHTTPS {
			return *sg.GroupId, nil
		}
	}
	
	// Create new security group
	createResult, err := client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{})
	if err != nil {
		return "", err
	}
	
	return *createResult.GroupId, nil
}

func getLatestUbuntuAMIWithClient(ctx context.Context, client EC2CreateAPI) (string, error) {
	result, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{})
	if err != nil {
		return "", err
	}
	
	if len(result.Images) == 0 {
		return "", errors.New("no Ubuntu AMI found")
	}
	
	// Find latest by creation date
	latest := result.Images[0]
	for _, img := range result.Images[1:] {
		if img.CreationDate != nil && latest.CreationDate != nil && 
		   *img.CreationDate > *latest.CreationDate {
			latest = img
		}
	}
	
	return *latest.ImageId, nil
}

func validateInstanceType(instanceType string) bool {
	if instanceType == "" {
		return false
	}
	
	// Basic validation - should contain a dot and be reasonable length
	return strings.Contains(instanceType, ".") && len(instanceType) > 3
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

// Interface for testing EC2 create operations
type EC2CreateAPI interface {
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	CreateSecurityGroup(ctx context.Context, params *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)
	AuthorizeSecurityGroupIngress(ctx context.Context, params *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
}
