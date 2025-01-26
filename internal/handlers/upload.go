package handlers

import (
"crypto/rand"
"encoding/hex"
"fmt"
"os"
"path/filepath"
"strconv"
"strings"
"time"

"fileditch/internal/db"

"github.com/gofiber/fiber/v2"
)

func generateRandomString(length int) (string, error) {
bytes := make([]byte, length/2)
if _, err := rand.Read(bytes); err != nil {
return "", err
}
return hex.EncodeToString(bytes), nil
}

func sanitizeFilename(filename string) string {
// Remove any path components
filename = filepath.Base(filename)
// Replace special characters
filename = strings.Map(func(r rune) rune {
if (r >= 'a' && r <= 'z') ||
(r >= 'A' && r <= 'Z') ||
(r >= '0' && r <= '9') ||
r == '-' || r == '_' || r == '.' {
return r
}
return '_'
}, filename)
return filename
}

func HandleUpload(c *fiber.Ctx) error {
// Parse multipart form
form, err := c.MultipartForm()
if err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "Invalid form data",
})
}

// Check for file
files := form.File["file"]
if len(files) == 0 {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "No file provided",
})
}
file := files[0]

// Get expiration time from form
expiresIn := 24 * 60 // Default to 24 hours (in minutes)
if expiresStr := c.FormValue("expireHours"); expiresStr != "" {
if hours, err := strconv.Atoi(expiresStr); err == nil {
expiresIn = hours * 60 // Convert hours to minutes
}
}

// Validate file size
maxSize := os.Getenv("MAX_FILE_SIZE")
if maxSize == "" {
maxSize = "10" // Default 10MB
}
var maxSizeMB int
fmt.Sscanf(maxSize, "%d", &maxSizeMB)
if file.Size > int64(maxSizeMB*1024*1024) {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": fmt.Sprintf("File too large. Maximum size is %dMB", maxSizeMB),
})
}

// Generate random string
randomLen := 32 // Default length
if envLen := os.Getenv("RANDOM_STRING_LENGTH"); envLen != "" {
fmt.Sscanf(envLen, "%d", &randomLen)
}
randomStr, err := generateRandomString(randomLen)
if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to generate file name",
})
}

// Create file name
origName := sanitizeFilename(file.Filename)
ext := filepath.Ext(origName)
baseName := strings.TrimSuffix(origName, ext)
newFileName := fmt.Sprintf("%s_%s%s", baseName, randomStr, ext)

// Save file
uploadDir := os.Getenv("UPLOAD_DIR")
if uploadDir == "" {
uploadDir = "uploads"
}
savePath := filepath.Join(uploadDir, newFileName)

if err := c.SaveFile(file, savePath); err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to save file",
})
}

// Save to database
expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Minute)
if err := db.SaveFile(origName, newFileName, expiresAt); err != nil {
// If database save fails, delete the uploaded file
os.Remove(savePath)
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to save file metadata",
})
}

// Return file URL
domain := os.Getenv("DOMAIN")
if domain == "" {
domain = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
}

fileURL := fmt.Sprintf("%s/file/%s", domain, newFileName)
return c.JSON(fiber.Map{
"url":       fileURL,
"expiresAt": expiresAt,
})
}