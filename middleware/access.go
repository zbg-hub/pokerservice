package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"pokerservice/errno"
)

type MyHandler func(c *gin.Context, ctx context.Context) errno.Payload

func Response(f MyHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		data := f(c, ctx)
		mkResp(c, data)
	}
}

func mkResp(c *gin.Context, data errno.Payload) {
	c.JSON(http.StatusOK, data)
}
