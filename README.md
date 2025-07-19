# Clouddley CLI

The Clouddley CLI is a command-line tool that allows you to interact with the Clouddley Platform from your terminal. With the CLI, you can perform various actions, such as deploying resources, managing configurations, and automating tasks.

## Installation

### MacOS, Linux, and WSL

Installing the latest version:

```bash
curl -L https://raw.githubusercontent.com/clouddley/cli/main/install.sh | sh
```

Installing a specific version:

```bash
curl -L https://raw.githubusercontent.com/clouddley/cli/main/install.sh | sh -s v0.1.4
```

### Windows

Run the following command in PowerShell:

```powershell
iwr https://raw.githubusercontent.com/clouddley/cli/main/install.ps1 -useb | iex
```

### Manual Installation

You can also download the binary for your platform from the [releases page](https://github.com/clouddley/cli/releases) and add it to your PATH.

## Usage

```bash
clouddley [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  triggr      Manage Triggr resources
  version     Print the version number of Clouddley CLI
  vm          Manage virtual machines across cloud providers

Flags:
  -h, --help   help for clouddley
```

### VM Management

The Clouddley CLI now supports creating and managing virtual machines on AWS:

#### AWS Prerequisites

1. Configure AWS credentials:
   ```bash
   # Set your AWS profile
   export AWS_PROFILE=your-profile-name
   
   # Or configure default credentials
   aws configure
   ```

2. Ensure you have SSH keys:
   ```bash
   # Generate an SSH key if you don't have one
   ssh-keygen -t ed25519 -C "your_email@example.com"
   # or
   ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
   ```

#### VM Commands

```bash
# Create a new AWS EC2 instance (interactive)
clouddley vm aws create

# List all instances created by Clouddley CLI
clouddley vm aws list

# Stop an instance
clouddley vm aws stop --id i-1234567890abcdef0

# Delete (terminate) an instance
clouddley vm aws delete --id i-1234567890abcdef0
```

#### Features

- **Interactive UI**: Uses Bubble Tea for beautiful interactive selection of environment and instance types
- **Smart SSH Key Management**: Automatically detects and imports local SSH keys to AWS
- **Environment-Based Pricing**: Shows different instance types for development/test vs production workloads
- **Cost Visibility**: Displays estimated monthly costs for each instance type
- **Safe Operations**: Confirmation prompts for destructive operations
- **AWS Profile Support**: Respects your AWS_PROFILE environment variable

## Contributing

If you have any suggestions, feature requests, or bug reports, please create an issue on the [GitHub repository](https://github.com/clouddley/cli).

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
