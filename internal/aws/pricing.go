package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
)

// PricingInfo holds pricing information for an instance type
type PricingInfo struct {
	InstanceType    string
	OnDemandPrice   float64
	EBSPrice        float64
	TotalPrice      float64
	FormattedPrice  string
}

// GetInstancePricing fetches pricing for a specific EC2 instance type using the default region
func GetInstancePricing(ctx context.Context, instanceType string) (*PricingInfo, error) {
	// Get the default region from AWS config
	cfg, err := GetAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}
	region := cfg.Region

	client, err := GetPricingClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing client: %w", err)
	}

	// Get EC2 instance pricing
	onDemandPrice, err := getEC2OnDemandPrice(ctx, client, instanceType, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get EC2 pricing: %w", err)
	}

	// Get EBS gp3 pricing for 100GB
	ebsPrice, err := getEBSPrice(ctx, client, region, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get EBS pricing: %w", err)
	}

	totalPrice := onDemandPrice + ebsPrice
	
	return &PricingInfo{
		InstanceType:   instanceType,
		OnDemandPrice:  onDemandPrice,
		EBSPrice:       ebsPrice,
		TotalPrice:     totalPrice,
		FormattedPrice: fmt.Sprintf("$%.2f/month", totalPrice),
	}, nil
}

// getEC2OnDemandPrice fetches the on-demand price for an EC2 instance
func getEC2OnDemandPrice(ctx context.Context, client *pricing.Client, instanceType, region string) (float64, error) {
	input := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters: []types.Filter{
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("instanceType"),
				Value: aws.String(instanceType),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("location"),
				Value: aws.String(getLocationName(region)),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("tenancy"),
				Value: aws.String("Shared"),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("operatingSystem"),
				Value: aws.String("Linux"),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("preInstalledSw"),
				Value: aws.String("NA"),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("capacitystatus"),
				Value: aws.String("Used"),
			},
		},
		MaxResults: aws.Int32(1),
	}

	result, err := client.GetProducts(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get products: %w", err)
	}

	if len(result.PriceList) == 0 {
		return 0, fmt.Errorf("no pricing found for instance type %s in %s", instanceType, region)
	}

	// Parse the pricing JSON
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(result.PriceList[0]), &product); err != nil {
		return 0, fmt.Errorf("failed to parse pricing JSON: %w", err)
	}

	// Navigate the complex pricing structure
	terms, ok := product["terms"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("terms not found in pricing data")
	}

	onDemand, ok := terms["OnDemand"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("OnDemand terms not found")
	}

	// Get the first (and usually only) on-demand offer
	for _, offer := range onDemand {
		offerMap, ok := offer.(map[string]interface{})
		if !ok {
			continue
		}

		priceDimensions, ok := offerMap["priceDimensions"].(map[string]interface{})
		if !ok {
			continue
		}

		// Get the first price dimension
		for _, dimension := range priceDimensions {
			dimMap, ok := dimension.(map[string]interface{})
			if !ok {
				continue
			}

			pricePerUnit, ok := dimMap["pricePerUnit"].(map[string]interface{})
			if !ok {
				continue
			}

			usdPrice, ok := pricePerUnit["USD"].(string)
			if !ok {
				continue
			}

			hourlyPrice, err := strconv.ParseFloat(usdPrice, 64)
			if err != nil {
				continue
			}

			// Convert hourly to monthly (24 hours * 30.44 days average)
			monthlyPrice := hourlyPrice * 24 * 30.44
			return monthlyPrice, nil
		}
	}

	return 0, fmt.Errorf("failed to extract price from pricing data")
}

// getEBSPrice fetches EBS gp3 pricing per GB
func getEBSPrice(ctx context.Context, client *pricing.Client, region string, sizeGB int) (float64, error) {
	input := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters: []types.Filter{
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("productFamily"),
				Value: aws.String("Storage"),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("volumeType"),
				Value: aws.String("General Purpose"),
			},
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("location"),
				Value: aws.String(getLocationName(region)),
			},
		},
		MaxResults: aws.Int32(10),
	}

	result, err := client.GetProducts(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get EBS products: %w", err)
	}

	for _, priceListItem := range result.PriceList {
		var product map[string]interface{}
		if err := json.Unmarshal([]byte(priceListItem), &product); err != nil {
			continue
		}

		// Check if this is gp3 storage
		productMap, ok := product["product"].(map[string]interface{})
		if !ok {
			continue
		}

		attributes, ok := productMap["attributes"].(map[string]interface{})
		if !ok {
			continue
		}

		storageMedia, ok := attributes["storageMedia"].(string)
		if !ok || !strings.Contains(strings.ToLower(storageMedia), "ssd") {
			continue
		}

		// Parse pricing
		terms, ok := product["terms"].(map[string]interface{})
		if !ok {
			continue
		}

		onDemand, ok := terms["OnDemand"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, offer := range onDemand {
			offerMap, ok := offer.(map[string]interface{})
			if !ok {
				continue
			}

			priceDimensions, ok := offerMap["priceDimensions"].(map[string]interface{})
			if !ok {
				continue
			}

			for _, dimension := range priceDimensions {
				dimMap, ok := dimension.(map[string]interface{})
				if !ok {
					continue
				}

				pricePerUnit, ok := dimMap["pricePerUnit"].(map[string]interface{})
				if !ok {
					continue
				}

				usdPrice, ok := pricePerUnit["USD"].(string)
				if !ok {
					continue
				}

				pricePerGBMonth, err := strconv.ParseFloat(usdPrice, 64)
				if err != nil {
					continue
				}

				// Calculate total EBS cost for the specified size
				totalEBSCost := pricePerGBMonth * float64(sizeGB)
				return totalEBSCost, nil
			}
		}
	}

	return 0, fmt.Errorf("no EBS pricing found for region %s", region)
}

// getLocationName converts AWS region to location name used in pricing API
func getLocationName(region string) string {
	locationMap := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-east-2":      "US East (Ohio)",
		"us-west-1":      "US West (N. California)",
		"us-west-2":      "US West (Oregon)",
		"af-south-1":     "Africa (Cape Town)",
		"ap-east-1":      "Asia Pacific (Hong Kong)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ap-south-2":     "Asia Pacific (Hyderabad)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-southeast-3": "Asia Pacific (Jakarta)",
		"ap-southeast-4": "Asia Pacific (Melbourne)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-northeast-3": "Asia Pacific (Osaka)",
		"ca-central-1":   "Canada (Central)",
		"eu-central-1":   "Europe (Frankfurt)",
		"eu-central-2":   "Europe (Zurich)",
		"eu-west-1":      "Europe (Ireland)",
		"eu-west-2":      "Europe (London)",
		"eu-west-3":      "Europe (Paris)",
		"eu-south-1":     "Europe (Milan)",
		"eu-south-2":     "Europe (Spain)",
		"eu-north-1":     "Europe (Stockholm)",
		"me-south-1":     "Middle East (Bahrain)",
		"me-central-1":   "Middle East (UAE)",
		"sa-east-1":      "South America (Sao Paulo)",
		"us-gov-west-1": "AWS GovCloud (US-West)",
		"us-gov-east-1": "AWS GovCloud (US-East)",
	}

	if location, ok := locationMap[region]; ok {
		return location
	}
	return region // Return the region itself if not found in map
}
