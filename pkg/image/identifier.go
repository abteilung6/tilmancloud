package image

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func GenerateImageIDFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}

	hashSum := hash.Sum(nil)
	shortHash := hex.EncodeToString(hashSum)[:16]

	return fmt.Sprintf("fedora-43-aarch64-%s", shortHash), nil
}
