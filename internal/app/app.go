package app

import (
	"eth_bal/configs"
	v1 "eth_bal/internal/contoller/http/v1"
	"eth_bal/internal/usecase"
	"eth_bal/pkg/httpserver"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func Run(cfg *configs.Config) error {
	checkerUseCase := usecase.New(cfg)
	handler := gin.New()
	v1.NewRouter(handler, checkerUseCase)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		fmt.Printf("app - Run - signal: %s\n", s.String())
	case err := <-httpServer.Notify():
		fmt.Printf("app - Run - httpServer.Notify: %v\n", err)
	}

	if err := httpServer.Shutdown(); err != nil {
		fmt.Printf("app - Run - httpServer.Shutdown: %v\n", err)
	}
	return nil
}
