package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure appropriately for production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.New().String()
		},
	})
}

// Logger middleware for structured logging
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log as structured JSON
		fields := logrus.Fields{
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"status":      param.StatusCode,
			"latency":     param.Latency.String(),
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
			"request_id":  param.Keys["X-Request-ID"],
		}

		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
		}

		entry := logger.WithFields(fields)
		if param.StatusCode >= 400 {
			entry.Error("HTTP request")
		} else {
			entry.Info("HTTP request")
		}

		return "" // Don't return anything as we're logging directly
	})
}

// RateLimit middleware for request rate limiting
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use Redis or similar
	requests := make(map[string][]time.Time)
	
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// Clean old requests
		if clientRequests, exists := requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range clientRequests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			requests[clientIP] = validRequests
		}
		
		// Check rate limit
		if len(requests[clientIP]) >= requestsPerMinute {
			c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(time.Minute).Unix(), 10))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		
		// Add current request
		requests[clientIP] = append(requests[clientIP], now)
		
		c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(requestsPerMinute-len(requests[clientIP])))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(time.Minute).Unix(), 10))
		
		c.Next()
	})
}

// Auth middleware for API authentication
func Auth(secretKey string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip auth for health check and public endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/health") ||
		   strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Validate token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token required",
				"code":  "TOKEN_REQUIRED",
			})
			c.Abort()
			return
		}

		// Simple token validation - in production, use proper JWT validation
		if !isValidToken(token, secretKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Set user context (extract from token in real implementation)
		c.Set("user_id", "default-user")
		c.Next()
	})
}

// Metrics middleware for collecting request metrics
func Metrics() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		// Collect metrics
		duration := time.Since(start)
		
		// In a real implementation, send metrics to Prometheus or similar
		// For now, just log them
		logrus.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration.String(),
			"user_agent": c.Request.UserAgent(),
		}).Info("Request metrics")
	})
}

// isValidToken validates the authentication token
func isValidToken(token, secretKey string) bool {
	// Simple validation - in production, implement proper JWT validation
	// For now, accept any non-empty token
	return strings.TrimSpace(token) != ""
}

// RequestSizeLimit middleware to limit request body size
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
				"code":  "REQUEST_TOO_LARGE",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// Timeout middleware to set request timeout
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
}