package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveFile saves a file to the public/upload directory with year/month structure
// and returns the relative path to the file
func SaveFile(file multipart.File, fileName string) (string, error) {
	// Get current year and month
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// Create directory structure: public/upload/year/month
	uploadDir := filepath.Join("public", "upload", year, month)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Generate a unique filename
	uniqueFileName := generateUniqueFileName(fileName)

	// Ensure the file extension is preserved
	ext := GetFileExtension(fileName)
	if ext != "" && !strings.HasSuffix(uniqueFileName, "."+ext) {
		uniqueFileName = uniqueFileName + "." + ext
	}

	// Full path to save the file
	fullPath := filepath.Join(uploadDir, uniqueFileName)

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return the relative path from public directory
	relativePath := filepath.Join("upload", year, month, uniqueFileName)
	return relativePath, nil
}

// generateUniqueFileName generates a unique filename by adding a random prefix
func generateUniqueFileName(originalName string) string {
	// Generate random bytes for uniqueness
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp if random generation fails
		return fmt.Sprintf("%d_%s", time.Now().Unix(), originalName)
	}

	// Convert to hex string
	randomString := hex.EncodeToString(randomBytes)

	// Combine with original filename
	return fmt.Sprintf("%s_%s", randomString, originalName)
}

// GetFileExtension extracts the file extension from a filename
func GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// SaveFileFromBytes saves a file from byte array to the public/upload directory
// and returns the relative path to the file
func SaveFileFromBytes(fileBytes []byte, fileName string) (string, error) {
	// Get current year and month
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// Create directory structure: public/upload/year/month
	uploadDir := filepath.Join("public", "upload", year, month)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// If no filename provided, use a default one
	if fileName == "" {
		fileName = "upload"
	}

	// Generate a unique filename with extension
	uniqueFileName := generateUniqueFileName(fileName)

	// Ensure the file extension is preserved
	ext := GetFileExtension(fileName)
	if ext != "" && !strings.HasSuffix(uniqueFileName, "."+ext) {
		uniqueFileName = uniqueFileName + "." + ext
	}

	// Full path to save the file
	fullPath := filepath.Join(uploadDir, uniqueFileName)

	// Write the file
	if err := os.WriteFile(fullPath, fileBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return the relative path from public directory
	relativePath := filepath.Join("upload", year, month, uniqueFileName)
	return relativePath, nil
}
