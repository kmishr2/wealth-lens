package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
)

const maxRateLimitClients = 10000

type rateWindow struct {
	started time.Time
	count   int
}

// RateLimit applies a fixed-window limit per client IP. It is deliberately
// process-local; replace it with a shared store before horizontally scaling.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]rateWindow)
	now := time.Now

	return func(c *gin.Context) {
		current := now()
		key := c.ClientIP()
		mu.Lock()
		if _, exists := clients[key]; !exists && len(clients) >= maxRateLimitClients {
			for client, candidate := range clients {
				if current.Sub(candidate.started) >= window {
					delete(clients, client)
				}
			}
			if len(clients) >= maxRateLimitClients {
				key = "__overflow__"
			}
		}
		entry, ok := clients[key]
		if !ok || current.Sub(entry.started) >= window {
			entry = rateWindow{started: current}
		}
		entry.count++
		clients[key] = entry
		remaining := limit - entry.count
		if remaining < 0 {
			remaining = 0
		}
		retryAfter := int(time.Until(entry.started.Add(window)).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		mu.Unlock()

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if entry.count > limit {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			common.RespondError(c, common.TooManyRequests("Too many authentication requests"))
			c.Abort()
			return
		}
		c.Next()
	}
}
