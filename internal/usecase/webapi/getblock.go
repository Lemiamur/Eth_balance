package webapi

import (
	"encoding/json"
	"errors"
	"eth_bal/internal/models"
	"eth_bal/internal/util"
	"eth_bal/pkg/jsonrpc"
	"eth_bal/pkg/log"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	_attempts = 5
	_delay    = 1 * time.Second
)

func GetLatestBlockNumber(client *http.Client, apiKey string) (string, error) {
	var result string
	err := util.RetryWithBackoff(_attempts, _delay, func() error {
		request := models.JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "eth_blockNumber",
			Params:  []any{},
			ID:      1,
		}
		var response models.JSONRPCResponse
		start := time.Now()
		if err := jsonrpc.SendJSONRPCRequest(client, apiKey, request, &response); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"method":  "eth_blockNumber",
				"attempt": "retried",
				"error":   err.Error(),
			}).Error("Failed to fetch block number")
			return err
		}
		duration := time.Since(start)
		log.Logger.WithFields(logrus.Fields{
			"method":   "eth_blockNumber",
			"duration": duration,
		}).Info("Block number fetched successfully")

		if response.Error != nil {
			log.Logger.WithFields(logrus.Fields{
				"method":  "eth_blockNumber",
				"code":    response.Error.Code,
				"message": response.Error.Message,
			}).Warn("API responded with error")
			return errors.New(response.Error.Message)
		}
		result = util.TrimQuotes(string(response.Result))
		return nil
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

func GetBlocksByNumbers(client *http.Client, apiKey string, blockNumbers []string, fullTx bool) ([]*models.Block, error) {
	var blocks []*models.Block
	err := util.RetryWithBackoff(_attempts, _delay, func() error {
		requests := make([]models.JSONRPCRequest, len(blockNumbers))
		for i, blockNumber := range blockNumbers {
			requests[i] = models.JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "eth_getBlockByNumber",
				Params:  []any{blockNumber, fullTx},
				ID:      int64(i + 1),
			}
		}

		var responses []models.JSONRPCResponse
		if err := jsonrpc.SendBatchJSONRPCRequest(client, apiKey, requests, &responses); err != nil {
			return err
		}

		blocks = make([]*models.Block, len(responses))
		var wg sync.WaitGroup
		for i, response := range responses {
			wg.Add(1)
			go func(i int, response models.JSONRPCResponse) {
				defer wg.Done()
				var block models.Block
				if err := json.Unmarshal(response.Result, &block); err != nil {
					blocks = nil
					return
				}
				blocks[i] = &block
			}(i, response)
		}
		wg.Wait()

		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}
