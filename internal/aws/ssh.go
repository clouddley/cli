package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// SSHKeyInfo represents SSH key information
type SSHKeyInfo struct {
	Path string
	Type string // "rsa" or "ed25519"
}

// CheckLocalSSHKeys checks for existing SSH keys and returns available options
func CheckLocalSSHKeys() ([]SSHKeyInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	var keys []SSHKeyInfo

	// Check for RSA key
	rsaPath := filepath.Join(sshDir, "id_rsa.pub")
	if _, err := os.Stat(rsaPath); err == nil {
		keys = append(keys, SSHKeyInfo{
			Path: rsaPath,
			Type: "rsa",
		})
	}

	// Check for Ed25519 key
	ed25519Path := filepath.Join(sshDir, "id_ed25519.pub")
	if _, err := os.Stat(ed25519Path); err == nil {
		keys = append(keys, SSHKeyInfo{
			Path: ed25519Path,
			Type: "ed25519",
		})
	}

	return keys, nil
}

// ReadSSHPublicKey reads the SSH public key from the given path
func ReadSSHPublicKey(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH key from %s: %w", path, err)
	}

	return string(content), nil
}

// CheckAWSKeyPair checks if the clouddley-default-key exists in AWS
func CheckAWSKeyPair(ctx context.Context, client *ec2.Client) (bool, error) {
	input := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{"clouddley-default-key"},
	}

	_, err := client.DescribeKeyPairs(ctx, input)
	if err != nil {
		// Check if it's a "not found" error by looking at the error message
		errMsg := fmt.Sprintf("%v", err)
		if strings.Contains(errMsg, "InvalidKeyPair.NotFound") || 
		   strings.Contains(errMsg, "does not exist") {
			// Key pair doesn't exist
			return false, nil
		}
		return false, fmt.Errorf("failed to check key pair: %w", err)
	}

	return true, nil
}

// ImportSSHKeyPair imports the SSH public key to AWS
func ImportSSHKeyPair(ctx context.Context, client *ec2.Client, publicKeyContent string) error {
	input := &ec2.ImportKeyPairInput{
		KeyName:           &[]string{"clouddley-default-key"}[0],
		PublicKeyMaterial: []byte(publicKeyContent),
	}

	_, err := client.ImportKeyPair(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to import key pair: %w", err)
	}

	return nil
}
