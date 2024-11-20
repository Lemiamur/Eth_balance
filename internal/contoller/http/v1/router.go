package v1

import (
	"eth_bal/internal/usecase"
	"eth_bal/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(handler *gin.Engine, t usecase.CheckBlock) {
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLE")
	handler.GET("/swagger/*any", swaggerHandler)

	handler.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := handler.Group("/v1")
	{
		newEthCheckRoutes(api, t)
	}
}

func newEthCheckRoutes(router *gin.RouterGroup, t usecase.CheckBlock) {
	router.GET("/check", func(c *gin.Context) {
		result := t.Check()
		log.Logger.WithField("result", result).Info("Sending response")
		c.JSON(http.StatusOK, result)
	})

}
