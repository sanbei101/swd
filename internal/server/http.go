package server

import (
	"swd-new/internal/handler"
	"swd-new/internal/middleware"
	"swd-new/pkg/helper/resp"
	"swd-new/pkg/log"

	"github.com/gin-gonic/gin"
)

func NewServerHTTP(
	logger *log.Logger,
	sensitiveCheckHandler *handler.SensitiveCheckHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(
		middleware.CORSMiddleware(),
	)
	r.GET("/", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{
			"say": "Hi Nunu!",
		})
	})
	r.POST("/sensitive_check", sensitiveCheckHandler.Check)
	r.GET("/word_manage", sensitiveCheckHandler.ListWords)
	r.POST("/word_manage", sensitiveCheckHandler.CreateWord)
	r.PUT("/word_manage/:id", sensitiveCheckHandler.UpdateWord)
	r.DELETE("/word_manage/:id", sensitiveCheckHandler.DeleteWord)

	return r
}
