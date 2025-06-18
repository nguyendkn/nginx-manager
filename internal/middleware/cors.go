package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/configs"
)

// CORSMiddleware creates a CORS middleware using environment configuration
func CORSMiddleware(env *configs.Environment) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers based on environment configuration
		origins := env.GetCORSAllowedOrigins()
		methods := env.GetCORSAllowedMethods()
		headers := env.GetCORSAllowedHeaders()

		// Handle allowed origins
		origin := c.Request.Header.Get("Origin")
		if len(origins) > 0 {
			if origins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				// Check if origin is in allowed list
				for _, allowedOrigin := range origins {
					if allowedOrigin == origin {
						c.Header("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
		}

		// Set allowed methods
		if len(methods) > 0 {
			methodsStr := ""
			for i, method := range methods {
				if i > 0 {
					methodsStr += ", "
				}
				methodsStr += method
			}
			c.Header("Access-Control-Allow-Methods", methodsStr)
		}

		// Set allowed headers
		if len(headers) > 0 {
			if headers[0] == "*" {
				c.Header("Access-Control-Allow-Headers", "*")
			} else {
				headersStr := ""
				for i, header := range headers {
					if i > 0 {
						headersStr += ", "
					}
					headersStr += header
				}
				c.Header("Access-Control-Allow-Headers", headersStr)
			}
		}

		// Set other CORS headers
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
