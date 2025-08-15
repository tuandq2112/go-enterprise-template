package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-clean-ddd-es-template/pkg/auth"

	"github.com/spf13/cobra"
)

var generateKeysCmd = &cobra.Command{
	Use:   "generate-keys",
	Short: "Generate RSA key pair for JWT signing",
	Long:  `Generate RSA private and public keys for JWT token signing and verification`,
	Run: func(cmd *cobra.Command, args []string) {
		generateRSAKeys()
	},
}

func init() {
	rootCmd.AddCommand(generateKeysCmd)
}

func generateRSAKeys() {
	// Generate RSA key pair (2048 bits for security)
	privateKey, publicKey, err := auth.GenerateRSAKeyPair(2048)
	if err != nil {
		fmt.Printf("Failed to generate RSA key pair: %v\n", err)
		os.Exit(1)
	}

	// Export keys to PEM format
	privateKeyPEM := auth.ExportPrivateKeyPEM(privateKey)
	publicKeyPEM := auth.ExportPublicKeyPEM(publicKey)

	// Create keys directory
	keysDir := "./keys"
	if err := os.MkdirAll(keysDir, 0o755); err != nil {
		fmt.Printf("Failed to create keys directory: %v\n", err)
		os.Exit(1)
	}

	// Write private key to file
	privateKeyPath := filepath.Join(keysDir, "private.pem")
	if err := os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0o600); err != nil {
		fmt.Printf("Failed to write private key: %v\n", err)
		os.Exit(1)
	}

	// Write public key to file
	publicKeyPath := filepath.Join(keysDir, "public.pem")
	if err := os.WriteFile(publicKeyPath, []byte(publicKeyPEM), 0o644); err != nil {
		fmt.Printf("Failed to write public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("RSA key pair generated successfully!\n")
	fmt.Printf("Private key: %s\n", privateKeyPath)
	fmt.Printf("Public key: %s\n", publicKeyPath)
	fmt.Printf("\n⚠️  IMPORTANT: Keep your private key secure and never commit it to version control!\n")
}
