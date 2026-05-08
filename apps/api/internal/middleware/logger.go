package middleware

import (
	"bytes"
	"io"
	"log"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		log.Printf("[REQ] %s %s | query: %s | body: %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
			string(bodyBytes),
		)

		c.Next()
	}
}
