package middleware

import (
"os"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/fiber/v2/middleware/session"
)

const AuthenticatedKey = "authenticated"

func NewAuthMiddleware(store *session.Store) fiber.Handler {
return func(c *fiber.Ctx) error {
// Skip auth check for public file access and static files
if c.Path() == "/" || c.Path() == "/login" || c.Path() == "/file/" {
return c.Next()
}

// Get session
sess, err := store.Get(c)
if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to get session",
})
}

// Check if user is authenticated
auth := sess.Get(AuthenticatedKey)
if auth == nil {
return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
"error": "Unauthorized",
})
}

return c.Next()
}
}

func ValidatePassword(password string) bool {
sitePassword := os.Getenv("SITE_PASSWORD")
if sitePassword == "" {
sitePassword = "admin123" // Default password, should be changed in production
}
return password == sitePassword
}

func Login(store *session.Store) fiber.Handler {
return func(c *fiber.Ctx) error {
var body struct {
Password string `json:"password"`
}

if err := c.BodyParser(&body); err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "Invalid request body",
})
}

if !ValidatePassword(body.Password) {
return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
"error": "Invalid password",
})
}

sess, err := store.Get(c)
if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to create session",
})
}

sess.Set(AuthenticatedKey, true)
if err := sess.Save(); err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to save session",
})
}

// Return success JSON response
return c.JSON(fiber.Map{
"success": true,
"message": "Login successful",
})
}
}

func Logout(store *session.Store) fiber.Handler {
return func(c *fiber.Ctx) error {
sess, err := store.Get(c)
if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to get session",
})
}

sess.Delete(AuthenticatedKey)
if err := sess.Save(); err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "Failed to save session",
})
}

// Return success JSON response
return c.JSON(fiber.Map{
"success": true,
"message": "Logout successful",
})
}
}