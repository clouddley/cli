package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestInstallCmd(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tempDir, err := os.MkdirTemp("", "clouddley-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("HOME", tempDir)

	authKeyFile := filepath.Join(tempDir, ".ssh", "authorized_keys")
	err = os.MkdirAll(filepath.Dir(authKeyFile), 0700)
	if err != nil {
		t.Fatalf("Error creating .ssh directory: %v", err)
	}

	cmd := &cobra.Command{
		Use: "clouddley",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(triggrCmd)
	cmd.SetArgs([]string{"triggr", "install"})

	err = installCmd.Execute()
	if err != nil {
		t.Fatalf("Error executing installCmd: %v", err)
	}

	content, err := os.ReadFile(authKeyFile)
	if err != nil {
		t.Fatalf("Error reading authorized_keys file: %v", err)
	}

	expectedKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCYDppn"
	if !strings.Contains(string(content), expectedKey) {
		t.Fatalf("Expected public key not found in authorized_keys file")
	}
}
