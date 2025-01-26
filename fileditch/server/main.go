package main

import (
"fmt"
"log"
"os"
"time"

"fileditch/internal/db"
"fileditch/internal/handlers"
"fileditch/internal/middleware"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/fiber/v2/middleware/cors"
"github.com/gofiber/fiber/v2/middleware/helmet"
"github.com/gofiber/fiber/v2/middleware/limiter"
"github.com/gofiber/fiber/v2/middleware/logger"
"github.com/gofiber/fiber/v2/middleware/session"
_ "github.com/joho/godotenv/autoload"
)

func main() {
// Create required directories
requiredDirs := []string{"uploads", "data"}
for _, dir := range requiredDirs {
if err := os.MkdirAll(dir, 0755); err != nil {
log.Fatalf("Failed to create directory %s: %v", dir, err)
}
}

// Initialize database
if err := db.Initialize(); err != nil {
log.Fatalf("Failed to initialize database: %v", err)
}
defer db.Close()

// Initialize fiber app
app := fiber.New(fiber.Config{
BodyLimit: getMaxFileSize(),
ErrorHandler: func(c *fiber.Ctx, err error) error {
statusCode := fiber.StatusInternalServerError
if e, ok := err.(*fiber.Error); ok {
statusCode = e.Code
}
return c.Status(statusCode).JSON(fiber.Map{
"error": err.Error(),
})
},
})

// Add middleware
app.Use(logger.New())
app.Use(cors.New())
app.Use(helmet.New())

// Rate limiting
app.Use(limiter.New(limiter.Config{
Max:        100,
Expiration: 5 * time.Minute,
KeyGenerator: func(c *fiber.Ctx) string {
return c.IP() // Rate limit by IP address
},
}))

// Initialize session store
store := session.New(session.Config{
Expiration:   24 * time.Hour,
CookieSecure: true,
})

// Start cleanup job
middleware.StartCleanupJob()

// Public routes
app.Static("/", "public") // Serve static files

// Route for root path - Show login or upload page based on auth status
app.Get("/", func(c *fiber.Ctx) error {
sess, err := store.Get(c)
if err != nil {
return c.SendFile("public/login.html")
}

auth := sess.Get(middleware.AuthenticatedKey)
if auth == nil {
return c.SendFile("public/login.html")
}
return c.SendFile("public/upload.html")
})

app.Get("/upload", func(c *fiber.Ctx) error {
sess, err := store.Get(c)
if err != nil {
return c.Redirect("/")
}

auth := sess.Get(middleware.AuthenticatedKey)
if auth == nil {
return c.Redirect("/")
}
return c.SendFile("public/upload.html")
})

app.Post("/login", middleware.Login(store))
app.Post("/logout", middleware.Logout(store))
app.Get("/file/:filename", handlers.HandleFileDownload)
app.Get("/file/info/:filename", handlers.HandleFileInfo)

// Create a route group for protected routes
api := app.Group("/")
api.Use(middleware.NewAuthMiddleware(store))
api.Post("/upload", handlers.HandleUpload)

// Error handler
app.Use(func(c *fiber.Ctx) error {
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
"error": "Not found",
})
})

// Start server
port := os.Getenv("PORT")
if port == "" {
port = "3000"
}

log.Printf("Server starting on port %s", port)
if err := app.Listen(":" + port); err != nil {
log.Fatalf("Error starting server: %v", err)
}
}

func getMaxFileSize() int {
maxSize := os.Getenv("MAX_FILE_SIZE")
var sizeInMB int
if _, err := fmt.Sscanf(maxSize, "%d", &sizeInMB); err != nil {
sizeInMB = 10 // Default 10MB
}
return sizeInMB * 1024 * 1024 // Convert to bytes
}
