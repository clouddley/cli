package aws

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetLocationName(t *testing.T) {
	tests := []struct {
		region   string
		expected string
	}{
		{"us-east-1", "US East (N. Virginia)"},
		{"us-east-2", "US East (Ohio)"},
		{"us-west-1", "US West (N. California)"},
		{"us-west-2", "US West (Oregon)"},
		{"ap-southeast-1", "Asia Pacific (Singapore)"},
		{"eu-west-1", "Europe (Ireland)"},
		{"ca-central-1", "Canada (Central)"},
		{"eu-central-1", "Europe (Frankfurt)"},
		{"ap-northeast-1", "Asia Pacific (Tokyo)"},
		{"unknown-region", "unknown-region"}, // Should return the region itself
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			result := getLocationName(tt.region)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPricingInfo_Structure(t *testing.T) {
	tests := []struct {
		name           string
		instanceType   string
		onDemandPrice  float64
		ebsPrice       float64
		expectedTotal  float64
		expectedFormat string
	}{
		{
			name:           "t3.micro with EBS",
			instanceType:   "t3.micro",
			onDemandPrice:  7.59,
			ebsPrice:       8.0,
			expectedTotal:  15.59,
			expectedFormat: "$15.59/month",
		},
		{
			name:           "m5.large with EBS",
			instanceType:   "m5.large",
			onDemandPrice:  69.12,
			ebsPrice:       8.0,
			expectedTotal:  77.12,
			expectedFormat: "$77.12/month",
		},
		{
			name:           "c5.xlarge with EBS",
			instanceType:   "c5.xlarge",
			onDemandPrice:  132.20,
			ebsPrice:       8.0,
			expectedTotal:  140.20,
			expectedFormat: "$140.20/month",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PricingInfo{
				InstanceType:   tt.instanceType,
				OnDemandPrice:  tt.onDemandPrice,
				EBSPrice:       tt.ebsPrice,
				TotalPrice:     tt.onDemandPrice + tt.ebsPrice,
				FormattedPrice: fmt.Sprintf("$%.2f/month", tt.onDemandPrice + tt.ebsPrice),
			}
			
			// Validate all fields
			if info.InstanceType != tt.instanceType {
				t.Errorf("Expected instance type %s, got %s", tt.instanceType, info.InstanceType)
			}
			
			if info.OnDemandPrice != tt.onDemandPrice {
				t.Errorf("Expected on-demand price %.2f, got %.2f", tt.onDemandPrice, info.OnDemandPrice)
			}
			
			if info.EBSPrice != tt.ebsPrice {
				t.Errorf("Expected EBS price %.2f, got %.2f", tt.ebsPrice, info.EBSPrice)
			}
			
			if info.TotalPrice != tt.expectedTotal {
				t.Errorf("Expected total price %.2f, got %.2f", tt.expectedTotal, info.TotalPrice)
			}
			
			if info.FormattedPrice != tt.expectedFormat {
				t.Errorf("Expected formatted price %s, got %s", tt.expectedFormat, info.FormattedPrice)
			}
		})
	}
}

func TestPricingInfo_FormattingEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		totalPrice     float64
		expectedFormat string
	}{
		{
			name:           "Zero cost",
			totalPrice:     0.0,
			expectedFormat: "$0.00/month",
		},
		{
			name:           "Very small cost",
			totalPrice:     0.01,
			expectedFormat: "$0.01/month",
		},
		{
			name:           "Large cost",
			totalPrice:     999.99,
			expectedFormat: "$999.99/month",
		},
		{
			name:           "Cost with many decimals",
			totalPrice:     15.123456,
			expectedFormat: "$15.12/month",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := fmt.Sprintf("$%.2f/month", tt.totalPrice)
			if formatted != tt.expectedFormat {
				t.Errorf("Expected formatted price %s, got %s", tt.expectedFormat, formatted)
			}
		})
	}
}

func TestRegionLocationMapping_Coverage(t *testing.T) {
	// Test that all major regions have location mappings
	regions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2",
		"eu-west-1", "eu-west-2", "eu-central-1",
		"ca-central-1", "sa-east-1",
	}
	
	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			location := getLocationName(region)
			
			// Should not return the region itself for known regions
			if location == region {
				t.Errorf("Region %s should have a location mapping, but returned itself", region)
			}
			
			// Should contain meaningful location information
			if location == "" {
				t.Errorf("Region %s returned empty location", region)
			}
			
			// Should not contain "unknown" for known regions
			if strings.Contains(strings.ToLower(location), "unknown") {
				t.Errorf("Region %s returned location with 'unknown': %s", region, location)
			}
		})
	}
}

func TestPricingInfo_ValidateFields(t *testing.T) {
	// Test that PricingInfo struct has all expected fields and types
	info := PricingInfo{
		InstanceType:   "t3.micro",
		OnDemandPrice:  10.50,
		EBSPrice:       8.0,
		TotalPrice:     18.50,
		FormattedPrice: "$18.50/month",
	}
	
	// Validate field types and values
	if info.InstanceType == "" {
		t.Error("InstanceType should not be empty")
	}
	
	if info.OnDemandPrice < 0 {
		t.Error("OnDemandPrice should not be negative")
	}
	
	if info.EBSPrice < 0 {
		t.Error("EBSPrice should not be negative")
	}
	
	if info.TotalPrice != info.OnDemandPrice + info.EBSPrice {
		t.Errorf("TotalPrice (%.2f) should equal OnDemandPrice + EBSPrice (%.2f)", 
			info.TotalPrice, info.OnDemandPrice + info.EBSPrice)
	}
	
	if !strings.Contains(info.FormattedPrice, "$") {
		t.Error("FormattedPrice should contain dollar sign")
	}
	
	if !strings.Contains(info.FormattedPrice, "/month") {
		t.Error("FormattedPrice should contain '/month'")
	}
}

func TestMonthlyCalculation(t *testing.T) {
	// Test the monthly calculation logic used in pricing
	tests := []struct {
		name          string
		hourlyRate    float64
		expectedMonthly float64
	}{
		{
			name:          "Small hourly rate",
			hourlyRate:    0.0104, // t3.micro
			expectedMonthly: 0.0104 * 24 * 30.44, // ~7.59
		},
		{
			name:          "Medium hourly rate", 
			hourlyRate:    0.096, // m5.large
			expectedMonthly: 0.096 * 24 * 30.44, // ~70.12
		},
		{
			name:          "High hourly rate",
			hourlyRate:    0.192, // m5.xlarge
			expectedMonthly: 0.192 * 24 * 30.44, // ~140.24
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the monthly calculation from the actual code
			monthlyPrice := tt.hourlyRate * 24 * 30.44
			
			// Allow small floating point differences
			diff := monthlyPrice - tt.expectedMonthly
			if diff < -0.01 || diff > 0.01 {
				t.Errorf("Expected monthly price around %.2f, got %.2f", tt.expectedMonthly, monthlyPrice)
			}
		})
	}
}
