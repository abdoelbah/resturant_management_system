package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

// SendJSONResponse sends a JSON response with the given status code and data
func SendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json") // Set Content-Type to "application/json"
	w.WriteHeader(status)                              // Write the status code
	err := json.NewEncoder(w).Encode(data)             // Encode the data to JSON and write it to the response
	if err != nil {                                    // Check if there was an error during encoding
		http.Error(w, err.Error(), http.StatusInternalServerError) // If so, send an internal server error response
	}
}

// HandleError standardizes error handling by sending a JSON error response
func HandleError(w http.ResponseWriter, status int, message string) {
	SendJSONResponse(w, status, map[string]string{
		"massage": message,
	})
}

// SaveImageFile saves the uploaded image file to the specified directory with a new name
func SaveImageFile(file io.Reader, table string, filename string) (string, error) {
	// Create directory structure if it doesn't exist
	fullPath := filepath.Join("uploads", table)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return "", err
	}

	// Generate new filename
	randomNumber := rand.Intn(1000)
	timestamp := time.Now().Unix()
	ext := filepath.Ext(filename)
	newFileName := fmt.Sprintf("%s_%d_%d%s", filepath.Base(table), timestamp, randomNumber, ext)
	newFilePath := filepath.Join(fullPath, newFileName)

	// Save the file
	destFile, err := os.Create(newFilePath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		return "", err
	}

	// Return the full path including directory
	return newFilePath, nil
}

// HashPassword hashes a plaintext password using bcrypt
func HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashPassword), nil
}

// DeleteImageFile deletes the specified file from the filesystem
func DeleteImageFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}

func ErrorWithTrace(err error, errMesssage string) error {

	if err != nil {
		// Skip 1 level to get the caller of this function
		_, file, line, _ := runtime.Caller(1)
		return fmt.Errorf("%s:%d: %v %s", file, line, err, errMesssage)
	}
	return nil
}

func CheckPassword(hashedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err
}
