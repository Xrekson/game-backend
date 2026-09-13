package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xrekson/game-backend/pkg/redis"
)

func RateLimiterMiddleware(redisClient *redis.Client, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s:%d", clientIP, time.Now().Unix()/120)
		fmt.Println("Rate limit key: ", key)
		count, err := redisClient.Incr(c.Request.Context(), key).Result()
		if err == nil {
			if count == 1 {
				redisClient.Expire(c.Request.Context(), key, 130*time.Second)
			}
			if count > int64(requestsPerMinute) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate limit exceeded",
					"retry_after": 120,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
