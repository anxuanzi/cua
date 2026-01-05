package cua

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// init automatically loads environment variables from .env files.
// It searches in the current directory and parent directories.
func init() {
	_ = LoadEnv()
}

// LoadEnv loads environment variables from .env files.
// It searches for .env in the current directory and up to 3 parent directories.
// This is useful for development when API keys are stored in .env files.
//
// The function silently ignores missing .env files, making it safe to call
// in production where environment variables are set differently.
//
// Returns an error only if a .env file exists but cannot be read.
func LoadEnv() error {
	// Try current directory first
	if err := loadEnvFile(".env"); err == nil {
		return nil
	}

	// Try to find .env in parent directories (up to 3 levels)
	wd, err := os.Getwd()
	if err != nil {
		return nil // Can't get working directory, skip silently
	}

	dir := wd
	for i := 0; i < 3; i++ {
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent

		envPath := filepath.Join(dir, ".env")
		if err := loadEnvFile(envPath); err == nil {
			return nil
		}
	}

	return nil // No .env found, that's fine
}

// loadEnvFile attempts to load a specific .env file.
func loadEnvFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return godotenv.Load(path)
}
