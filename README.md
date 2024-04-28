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
curl -L https://raw.githubusercontent.com/clouddley/cli/main/install.sh | sh -s 0.1.1
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

Flags:
  -h, --help   help for clouddley
```

## Contributing

If you have any suggestions, feature requests, or bug reports, please create an issue on the [GitHub repository](https://github.com/clouddley/cli).

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
