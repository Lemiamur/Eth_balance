package app

import (
	"eth_bal/configs"
	"eth_bal/internal/service"
	"eth_bal/pkg/log"
)

// Функция Run запускает основное приложение
func Run(cfg *configs.Config) error {
	log.Logger.Info("Запуск сервиса проверки балансов")
	service.EthChecker(cfg)
	return nil
}
