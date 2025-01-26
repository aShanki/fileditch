package handlers

import (
"mime"
"os"
"path/filepath"
"time"

"fileditch/internal/db"

"github.com/gofiber/fiber/v2"
)

func HandleFileDownload(c *fiber.Ctx) error {
// Get filename from URL parameter
filename := c.Params("filename")
if filename == "" {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "No filename provided",
})
}

// Get file info from database
file, err := db.GetFile(filename)
if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to retrieve file information",
})
}

if file == nil {
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
"error": "File not found",
})
}

// Check if file has expired
if time.Now().After(file.ExpiresAt) {
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
"error": "File has expired",
})
}

// Get physical file path
uploadDir := os.Getenv("UPLOAD_DIR")
if uploadDir == "" {
uploadDir = "uploads"
}
filePath := filepath.Join(uploadDir, file.Path)

// Check if file exists on disk
if _, err := os.Stat(filePath); os.IsNotExist(err) {
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
"error": "File not found",
})
}

// Determine content type
ext := filepath.Ext(file.Name)
contentType := mime.TypeByExtension(ext)
if contentType == "" {
contentType = "application/octet-stream"
}

// Set headers for download
c.Set("Content-Type", contentType)
c.Set("Content-Disposition", "inline; filename="+file.Name)

// Send file
return c.SendFile(filePath)
}

func HandleFileInfo(c *fiber.Ctx) error {
// Get filename from URL parameter
filename := c.Params("filename")
if filename == "" {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "No filename provided",
})
}

// Get file info from database
file, err := db.GetFile(filename)
if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to retrieve file information",
})
}

if file == nil {
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
"error": "File not found",
})
}

// Check if file has expired
if time.Now().After(file.ExpiresAt) {
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
"error": "File has expired",
})
}

// Get file size
uploadDir := os.Getenv("UPLOAD_DIR")
if uploadDir == "" {
uploadDir = "uploads"
}
filePath := filepath.Join(uploadDir, file.Path)

fileInfo, err := os.Stat(filePath)
if err != nil {
if os.IsNotExist(err) {
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
"error": "File not found",
})
}
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to get file information",
})
}

return c.JSON(fiber.Map{
"name": file.Name,
"size": fileInfo.Size(),
"expiresAt": file.ExpiresAt,
"createdAt": file.CreatedAt,
})
}