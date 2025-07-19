package aws

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsinternal "github.com/clouddley/clouddley/internal/aws"
	"github.com/clouddley/clouddley/internal/log"
	"github.com/clouddley/clouddley/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new AWS EC2 instance",
	Long:  `Create a new AWS EC2 instance with interactive selection of environment and instance type.`,
	Run:   runCreate,
}

func runCreate(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Show banner
	fmt.Print(ui.ShowBanner())

	// Validate AWS credentials
	if err := awsinternal.ValidateAWSCredentials(ctx); err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error: %v", err)))
		return
	}

	// Check/handle SSH keys
	selectedKey, err := handleSSHKeys(ctx)
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error: %v", err)))
		return
	}

	// Interactive environment selection
	envModel := ui.NewEnvironmentModel()
	p := tea.NewProgram(envModel)
	m, err := p.Run()
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error running environment selection: %v", err)))
		return
	}

	envChoice := m.(ui.EnvironmentModel).Selected()
	if envChoice == -1 {
		fmt.Println("Operation cancelled")
		return
	}

	// Get instance types and pricing
	instances, err := getInstanceTypes(ctx, envChoice == 0) // 0 = dev/test, 1 = production
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error fetching instance types: %v", err)))
		return
	}

	// Interactive instance type selection
	title := "Development/Test Instances"
	if envChoice == 1 {
		title = "Production Instances"
	}
	instanceModel := ui.NewInstanceSelectionModel(instances, title)
	p = tea.NewProgram(instanceModel)
	m, err = p.Run()
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error running instance selection: %v", err)))
		return
	}

	instanceChoice := m.(ui.InstanceSelectionModel).Selected()
	if instanceChoice == -1 {
		fmt.Println("Operation cancelled")
		return
	}

	selectedInstance := instances[instanceChoice]
	
	// Create the instance
	fmt.Println("Creating instance...")
	instanceInfo, err := createInstance(ctx, selectedInstance.Type, selectedKey)
	if err != nil {
		fmt.Println(ui.FormatError(fmt.Sprintf("Error creating instance: %v", err)))
		return
	}

	// Output success in table format
	fmt.Println(ui.FormatOutput("✓ Success", "Instance created successfully!"))
	fmt.Println()
	
	// Create table for instance details
	instanceTable := ui.NewInstanceDetailsTable()
	instanceTable.AddRow("Instance ID", instanceInfo.InstanceID)
	instanceTable.AddRow("Name", instanceInfo.Name)
	instanceTable.AddRow("Region", instanceInfo.Region)
	instanceTable.AddRow("Public IP", instanceInfo.PublicIP)
	instanceTable.AddRow("Instance Type", selectedInstance.Type)
	instanceTable.AddRow("Username", "ubuntu")
	instanceTable.AddRow("SSH Port", "22")
	instanceTable.AddRow("SSH Command", fmt.Sprintf("ssh ubuntu@%s", instanceInfo.PublicIP))
	
	fmt.Println(instanceTable.Render())
	
	// Post-creation: Ask if user wants to install Clouddley public key
	fmt.Println()
	confirmModel := ui.NewConfirmationModel("Install Clouddley public key for dashboard access?")
	p = tea.NewProgram(confirmModel)
	m, err = p.Run()
	if err != nil {
		log.Error("Error running confirmation prompt", "error", err)
		return
	}

	confirmResult := m.(ui.ConfirmationModel)
	if !confirmResult.Answered() {
		fmt.Println("Operation cancelled")
		return
	}

	if confirmResult.Selected() {
		// User chose yes, install the key
		fmt.Println()
		fmt.Println("Installing Clouddley public key on VM...")
		
		err = installClouddleyKey(ctx, instanceInfo.PublicIP)
		if err != nil {
			log.Error("Failed to install Clouddley public key", "error", err)
			fmt.Println(ui.FormatError(fmt.Sprintf("Failed to install key: %v", err)))
		} else {
			log.Info("Clouddley public key installed successfully")
			fmt.Println(ui.FormatOutput("✓ Success", "Clouddley public key installed! Your VM is now registered for dashboard authentication."))
		}
	} else {
		fmt.Println("Skipped installing Clouddley public key.")
	}
}

func handleSSHKeys(ctx context.Context) (*awsinternal.SSHKeyInfo, error) {
	client, err := awsinternal.GetEC2Client(ctx)
	if err != nil {
		return nil, err
	}

	// Check if AWS key pair already exists
	keyExists, err := awsinternal.CheckAWSKeyPair(ctx, client)
	if err != nil {
		return nil, err
	}

	if keyExists {
		log.Info("Using existing clouddley-default-key from AWS")
		return nil, nil // No need to import, key already exists
	}

	// Check local SSH keys
	localKeys, err := awsinternal.CheckLocalSSHKeys()
	if err != nil {
		return nil, err
	}

	if len(localKeys) == 0 {
		return nil, fmt.Errorf(`no SSH public key found. Generate one using:
ssh-keygen -t ed25519 -C "your_email@example.com" (recommended)
or
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

See https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent for details.

Then retry the command`)
	}

	var selectedKey *awsinternal.SSHKeyInfo

	if len(localKeys) == 1 {
		selectedKey = &localKeys[0]
		fmt.Printf("Using SSH key: %s (%s)\n", selectedKey.Path, selectedKey.Type)
	} else {
		// Multiple keys found, let user choose
		keyChoices := make([]string, len(localKeys))
		for i, key := range localKeys {
			keyChoices[i] = fmt.Sprintf("%s (%s)", key.Path, key.Type)
		}

		keyModel := ui.NewSSHKeySelectionModel(keyChoices)
		p := tea.NewProgram(keyModel)
		m, err := p.Run()
		if err != nil {
			return nil, fmt.Errorf("error running key selection: %w", err)
		}

		keyChoice := m.(ui.SSHKeySelectionModel).Selected()
		if keyChoice == -1 {
			return nil, fmt.Errorf("operation cancelled")
		}

		selectedKey = &localKeys[keyChoice]
	}

	// Read and import the key
	publicKeyContent, err := awsinternal.ReadSSHPublicKey(selectedKey.Path)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Importing SSH key to AWS...\n")
	err = awsinternal.ImportSSHKeyPair(ctx, client, publicKeyContent)
	if err != nil {
		return nil, err
	}

	log.Info("SSH key imported successfully")
	return selectedKey, nil
}

func getInstanceTypes(ctx context.Context, isDev bool) ([]ui.InstanceType, error) {
	var instanceSpecs []struct {
		Type   string
		CPU    string
		Memory string
	}
	
	if isDev {
		instanceSpecs = []struct {
			Type   string
			CPU    string
			Memory string
		}{
			{"t3.nano", "2", "0.5 GB"},
			{"t3.micro", "2", "1 GB"},
			{"t3.small", "2", "2 GB"},
			{"t3.medium", "2", "4 GB"},
			{"t3.large", "2", "8 GB"},
			{"t3.xlarge", "4", "16 GB"},
			{"t3.2xlarge", "8", "32 GB"},
		}
	} else {
		instanceSpecs = []struct {
			Type   string
			CPU    string
			Memory string
		}{
			{"m5.large", "2", "8 GB"},
			{"m5.xlarge", "4", "16 GB"},
			{"m5.2xlarge", "8", "32 GB"},
			{"c5.large", "2", "4 GB"},
			{"c5.xlarge", "4", "8 GB"},
			{"c5.2xlarge", "8", "16 GB"},
		}
	}
	
	// Show loading spinner while fetching pricing
	fmt.Println()
	loadingModel := ui.NewLoadingModel("Fetching live pricing data...")
	
	// Channel to signal completion
	done := make(chan []ui.InstanceType, 1)
	errorChan := make(chan error, 1)
	
	// Start the pricing fetch in a goroutine
	go func() {
		var instances []ui.InstanceType
		
		for _, spec := range instanceSpecs {
			pricingInfo, err := awsinternal.GetInstancePricing(ctx, spec.Type)
			if err != nil {
				log.Error("Failed to get pricing for instance type", "instance", spec.Type, "error", err)
				continue
			}
			
			instances = append(instances, ui.InstanceType{
				Type:        spec.Type,
				VCPUs:       spec.CPU,
				Memory:      spec.Memory,
				Disk:        "100 GB",
				MonthlyCost: pricingInfo.FormattedPrice,
			})
			
			log.Debug("Got pricing for instance", "instance", spec.Type, "price", pricingInfo.FormattedPrice)
		}
		
		if len(instances) == 0 {
			errorChan <- fmt.Errorf("failed to fetch pricing for any instance types")
			return
		}
		
		log.Info("Successfully fetched pricing data", "instances", len(instances))
		done <- instances
	}()
	
	// Start the loading spinner
	p := tea.NewProgram(loadingModel)
	go func() {
		p.Run()
	}()
	
	// Wait for completion
	select {
	case instances := <-done:
		loadingModel.Finish()
		p.Quit()
		fmt.Print("\r\033[K") // Clear the spinner line
		return instances, nil
	case err := <-errorChan:
		loadingModel.Finish()
		p.Quit()
		fmt.Print("\r\033[K") // Clear the spinner line
		return nil, err
	}
}

type InstanceInfo struct {
	InstanceID string
	Name       string
	PublicIP   string
	Region     string
}

func createInstance(ctx context.Context, instanceType string, sshKey *awsinternal.SSHKeyInfo) (*InstanceInfo, error) {
	client, err := awsinternal.GetEC2Client(ctx)
	if err != nil {
		return nil, err
	}

	// Get the AWS config to extract region
	cfg, err := awsinternal.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Get default VPC and subnet
	vpc, subnet, err := getDefaultVPCAndSubnet(ctx, client)
	if err != nil {
		return nil, err
	}

	// Create security group if needed
	sgID, err := createOrGetSecurityGroup(ctx, client, vpc)
	if err != nil {
		return nil, err
	}

	// Get latest Ubuntu AMI
	amiID, err := getLatestUbuntuAMI(ctx, client)
	if err != nil {
		return nil, err
	}

	// Generate instance name
	instanceName := fmt.Sprintf("clouddley-vm-%d", time.Now().Unix())

	// Create instance
	runResult, err := client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: types.InstanceType(instanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      aws.String("clouddley-default-key"),
		SecurityGroupIds: []string{sgID},
		SubnetId:     aws.String(subnet),
		BlockDeviceMappings: []types.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"), // Root device for Ubuntu
				Ebs: &types.EbsBlockDevice{
					VolumeSize:          aws.Int32(100), // 100GB
					VolumeType:          types.VolumeTypeGp3,
					DeleteOnTermination: aws.Bool(true),
					Encrypted:           aws.Bool(true),
				},
			},
		},
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(instanceName),
					},
					{
						Key:   aws.String("CreatedBy"),
						Value: aws.String("Clouddley"),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	instanceID := *runResult.Instances[0].InstanceId

	// Wait for instance to be running with loading spinner
	fmt.Println()
	loadingModel := ui.NewLoadingModel("Waiting for instance to be running...")
	
	// Channel to signal completion
	done := make(chan error, 1)
	
	// Start the waiter in a goroutine
	go func() {
		waiter := ec2.NewInstanceRunningWaiter(client)
		err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}, 5*time.Minute)
		done <- err
	}()
	
	// Start the loading spinner
	p := tea.NewProgram(loadingModel)
	go func() {
		p.Run()
	}()
	
	// Wait for completion
	err = <-done
	loadingModel.Finish()
	p.Quit()
	
	fmt.Print("\r\033[K") // Clear the spinner line
	
	if err != nil {
		return nil, fmt.Errorf("failed waiting for instance to be running: %w", err)
	}

	// Get instance details
	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	instance := result.Reservations[0].Instances[0]
	publicIP := ""
	if instance.PublicIpAddress != nil {
		publicIP = *instance.PublicIpAddress
	}

	return &InstanceInfo{
		InstanceID: instanceID,
		Name:       instanceName,
		PublicIP:   publicIP,
		Region:     cfg.Region,
	}, nil
}

func getDefaultVPCAndSubnet(ctx context.Context, client *ec2.Client) (string, string, error) {
	// Get default VPC
	vpcResult, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("is-default"),
				Values: []string{"true"},
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to describe VPCs: %w", err)
	}

	if len(vpcResult.Vpcs) == 0 {
		return "", "", fmt.Errorf("no default VPC found")
	}

	vpcID := *vpcResult.Vpcs[0].VpcId

	// Get first available subnet in default VPC
	subnetResult, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
			{
				Name:   aws.String("default-for-az"),
				Values: []string{"true"},
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to describe subnets: %w", err)
	}

	if len(subnetResult.Subnets) == 0 {
		return "", "", fmt.Errorf("no default subnet found")
	}

	subnetID := *subnetResult.Subnets[0].SubnetId

	return vpcID, subnetID, nil
}

func createOrGetSecurityGroup(ctx context.Context, client *ec2.Client, vpcID string) (string, error) {
	sgName := "clouddley-default-sg"
	
	// Check if security group exists
	result, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []string{sgName},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe security groups: %w", err)
	}

	var sgID string
	var needsRules bool

	if len(result.SecurityGroups) > 0 {
		sgID = *result.SecurityGroups[0].GroupId
		
		// Check if all required ports are already open
		sg := result.SecurityGroups[0]
		hasSSH := false
		hasHTTP := false
		hasHTTPS := false
		
		for _, rule := range sg.IpPermissions {
			if rule.IpProtocol != nil && *rule.IpProtocol == "tcp" {
				if rule.FromPort != nil && rule.ToPort != nil {
					port := *rule.FromPort
					// Check if rule allows 0.0.0.0/0
					for _, ipRange := range rule.IpRanges {
						if ipRange.CidrIp != nil && *ipRange.CidrIp == "0.0.0.0/0" {
							switch port {
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
			}
		}
		
		needsRules = !hasSSH || !hasHTTP || !hasHTTPS
	} else {
		// Create security group
		createResult, err := client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
			GroupName:   aws.String(sgName),
			Description: aws.String("Clouddley CLI default security group"),
			VpcId:       aws.String(vpcID),
		})
		if err != nil {
			return "", fmt.Errorf("failed to create security group: %w", err)
		}
		sgID = *createResult.GroupId
		needsRules = true
	}

	// Add missing rules if needed
	if needsRules {
		// Build permissions for missing rules only
		var permissions []types.IpPermission
		
		// Get current rules to avoid duplicates
		currentSG, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
			GroupIds: []string{sgID},
		})
		if err != nil {
			return "", fmt.Errorf("failed to describe security group: %w", err)
		}
		
		hasSSH := false
		hasHTTP := false
		hasHTTPS := false
		
		if len(currentSG.SecurityGroups) > 0 {
			for _, rule := range currentSG.SecurityGroups[0].IpPermissions {
				if rule.IpProtocol != nil && *rule.IpProtocol == "tcp" && rule.FromPort != nil {
					port := *rule.FromPort
					for _, ipRange := range rule.IpRanges {
						if ipRange.CidrIp != nil && *ipRange.CidrIp == "0.0.0.0/0" {
							switch port {
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
			}
		}
		
		// Add SSH rule if missing
		if !hasSSH {
			permissions = append(permissions, types.IpPermission{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(22),
				ToPort:     aws.Int32(22),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("SSH access"),
					},
				},
			})
		}
		
		// Add HTTP rule if missing
		if !hasHTTP {
			permissions = append(permissions, types.IpPermission{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(80),
				ToPort:     aws.Int32(80),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("HTTP access"),
					},
				},
			})
		}
		
		// Add HTTPS rule if missing
		if !hasHTTPS {
			permissions = append(permissions, types.IpPermission{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(443),
				ToPort:     aws.Int32(443),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("HTTPS access"),
					},
				},
			})
		}
		
		// Only authorize if we have permissions to add
		if len(permissions) > 0 {
			_, err = client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:       aws.String(sgID),
				IpPermissions: permissions,
			})
			if err != nil {
				return "", fmt.Errorf("failed to add security group rules: %w", err)
			}
		}
	}

	return sgID, nil
}

func getLatestUbuntuAMI(ctx context.Context, client *ec2.Client) (string, error) {
	// Try Ubuntu 24.04 LTS first (Noble Numbat)
	patterns := []string{
		"ubuntu/images/hvm-ssd/ubuntu-noble-24.04-amd64-server-*",
		"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*", // Fallback to 22.04
	}

	for _, pattern := range patterns {
		result, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("name"),
					Values: []string{pattern},
				},
				{
					Name:   aws.String("owner-id"),
					Values: []string{"099720109477"}, // Canonical
				},
			},
			Owners: []string{"099720109477"},
		})
		if err != nil {
			continue
		}

		if len(result.Images) > 0 {
			// Sort by creation date and get latest
			latest := result.Images[0]
			for _, img := range result.Images[1:] {
				if img.CreationDate != nil && latest.CreationDate != nil && *img.CreationDate > *latest.CreationDate {
					latest = img
				}
			}
			return *latest.ImageId, nil
		}
	}

	return "", fmt.Errorf("no Ubuntu AMI found")
}

// installClouddleyKey installs the Clouddley public key on the VM via SSH
func installClouddleyKey(ctx context.Context, publicIP string) error {
	// The Clouddley triggr public key (same as in cmd/triggr.go)
	sshPublicKey := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCYDppnNM+F+GtFWaVJXsvobX/i/2uZuLch9386ZEyVtGE1QmRjRkvwEHwFM23STuzbtmqTrYjEnmv3Xkywk7wE0r+OoJxTBIwJP+scg9rAu//3N6CoAKH0Ra1XgdRj8QzqF/1mm4T/Pxtzz3JSpSKwpzW3GtU4NcHuaPAAHavCpahCnZqpPMU90FgRCS9lSmw0EPQcU8kxxeEpFjifip4JBBx/WQuh/8KkBAX/DnWSAO9ynGzPMvOvWPTtQMi7IA7Y8vRWeThfpC/fnU8Tub+99w5h2Y1TnWtUrM49ZMa9WSLtP/+4xKieQPObq0JuX6itNFuuwbb/WHLgOYeZqQTdSeMc6GlSkqniYiAUAv7olBUERHf7QkD7hPOlaw9S/0MCU8DcuujZG2i6UvIkQ60dikvsX8rCiPvfN4Nw1mWh0a1rf9vUxTyCCb+7hh1iPV6RwMx6T4nBjFNjBglHFkYIE5kevLyX2vREJJen+GfZO2GVcnHaNRHBvXZVVEbwt1xRWhAOS+FFtcKUNV+54JsKTaZUEYvfwe/KNjEeOxucljkiK9IYw0IGXB9dtueOTKcirLhpGE9t6LqDhWE05kr0fl/hmnT/g9fHeZDm4jOF71iHogsrZtU5pH8QtTNhaffMkW4EJc+4W0a+boE+/S5Xracbr7D1WBhGC2epXkUWHw== "clouddley-triggr-public-key"`

	log.Debug("Installing Clouddley public key", "host", publicIP)

	// SSH command to install the key
	sshCommand := fmt.Sprintf(`
		# Create .ssh directory if it doesn't exist
		mkdir -p ~/.ssh
		chmod 700 ~/.ssh
		
		# Add the public key to authorized_keys if not already present
		if ! grep -q "clouddley-triggr-public-key" ~/.ssh/authorized_keys 2>/dev/null; then
			echo '%s' >> ~/.ssh/authorized_keys
			chmod 600 ~/.ssh/authorized_keys
			echo "Clouddley public key installed successfully"
		else
			echo "Clouddley public key already exists"
		fi
	`, sshPublicKey)

	// Execute SSH command with timeout
	cmd := exec.CommandContext(ctx, "ssh", 
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=30",
		"-o", "ServerAliveInterval=5",
		"-o", "ServerAliveCountMax=3",
		fmt.Sprintf("ubuntu@%s", publicIP),
		sshCommand)

	log.Debug("Executing SSH command", "command", cmd.String())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SSH command failed: %w, output: %s", err, string(output))
	}

	log.Debug("SSH command output", "output", string(output))

	// Verify the key was installed by checking if it exists
	verifyCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=30",
		fmt.Sprintf("ubuntu@%s", publicIP),
		"grep -q 'clouddley-triggr-public-key' ~/.ssh/authorized_keys && echo 'KEY_FOUND' || echo 'KEY_NOT_FOUND'")

	verifyOutput, err := verifyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("verification command failed: %w, output: %s", err, string(verifyOutput))
	}

	if !strings.Contains(string(verifyOutput), "KEY_FOUND") {
		return fmt.Errorf("key installation verification failed: %s", string(verifyOutput))
	}

	log.Debug("Key installation verified successfully")
	return nil
}
