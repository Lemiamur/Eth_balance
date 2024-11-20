package usecase

import (
	"eth_bal/configs"
	"eth_bal/internal/models"
	"eth_bal/internal/service"
)

type CheckBlock interface {
	Check() models.ResultBlock
}

type checkblock struct {
	cfg *configs.Config
}

func New(cfg *configs.Config) CheckBlock {
	return &checkblock{cfg: cfg}
}

func (t *checkblock) Check() models.ResultBlock {
	return service.EthChecker(t.cfg)
}
