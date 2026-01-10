package auth

import (
	"net/http"

	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/gin-gonic/gin"
)

const (
	//CookieName is the name of the auth cookie
	CookieName = "rcnbuild_token"
	// UserContextKey is the key used to store user in gin context
	UserContextKey = "user"
)

// Middleware that requires a valid JWT
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		tokenString, err := c.Cookie(CookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := ValidateToken(tokenString)
		if err != nil {
			// Clear invalid cookie
			c.SetCookie(CookieName, "", -1, "/", "", false, true)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Fetch user from database
		user, err := database.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Store user in context for handlers to use
		c.Set(UserContextKey, user)
		c.Next()
	}
}

// Retrieve the authenticated user from context
func GetCurrentUser(c *gin.Context) *database.User {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	return user.(*database.User)
}

// Set the JWT cookie
func SetAuthCookie(c *gin.Context, token string) {
	// HTTP-only cookie prevents JavaScript access (XSS protection)
	// Secure=true in production (HTTPS only)
	// SameSite=Lax prevents CSRF while allowing normal navigation
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		CookieName,
		token,
		60*60*24*7, // 7 days in seconds
		"/",
		"",    // Domain (empty = current domain)
		false, // Secure (set true in production with HTTPS)
		true,  // HTTP-only
	)
}

// Remove the auth cookie
func ClearAuthCookie(c *gin.Context) {
	c.SetCookie(CookieName, "", -1, "/", "", false, true)
}
