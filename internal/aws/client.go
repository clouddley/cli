package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
)

// GetEC2Client returns an AWS EC2 client configured with the current AWS profile
func GetEC2Client(ctx context.Context) (*ec2.Client, error) {
	awsProfile := os.Getenv("AWS_PROFILE")
	
	var cfg aws.Config
	var err error
	
	if awsProfile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsProfile))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	return ec2.NewFromConfig(cfg), nil
}

// GetPricingClient returns an AWS Pricing client
func GetPricingClient(ctx context.Context) (*pricing.Client, error) {
	awsProfile := os.Getenv("AWS_PROFILE")
	
	var cfg aws.Config
	var err error
	
	if awsProfile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsProfile), config.WithRegion("us-east-1"))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	return pricing.NewFromConfig(cfg), nil
}

// GetAWSConfig returns the AWS configuration
func GetAWSConfig(ctx context.Context) (aws.Config, error) {
	awsProfile := os.Getenv("AWS_PROFILE")
	
	var cfg aws.Config
	var err error
	
	if awsProfile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsProfile))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
	}
	
	if err != nil {
		return cfg, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	return cfg, nil
}

// ValidateAWSCredentials checks if AWS credentials are properly configured
func ValidateAWSCredentials(ctx context.Context) error {
	client, err := GetEC2Client(ctx)
	if err != nil {
		return err
	}
	
	// Try a simple API call to validate credentials
	_, err = client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return fmt.Errorf("AWS credentials invalid or AWS_PROFILE not set. Set AWS_PROFILE (e.g., export AWS_PROFILE=corp) and ensure ~/.aws/credentials is configured: %w", err)
	}
	
	return nil
}
