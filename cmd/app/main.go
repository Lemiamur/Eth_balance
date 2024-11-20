package main

import (
	"eth_bal/configs"
	"eth_bal/internal/app"
	"eth_bal/internal/cache"
	"eth_bal/pkg/log"
	"fmt"
	"time"
)

func main() {
	startTime := time.Now()
	cfg, err := configs.LoadConfig("configs/config.yml")
	if err != nil {
		log.Logger.Errorf("Ошибка загрузки конфигурации: %v", err)
		return
	}

	cache.GetGlobalBlockCache(cfg.CacheSize)

	if err := app.Run(cfg); err != nil {
		log.Logger.Errorf("Ошибка запуска приложения: %v", err)
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Программа выполнялась: %v\n", duration)
}
