package api

import (
	"encoding/json"
	"errors"
	models "eth_bal/internal/models"
	util "eth_bal/internal/util"
	jsonrpc "eth_bal/pkg/jsonrpc"
	"eth_bal/pkg/log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func GetLatestBlockNumber(client *http.Client, apiKey string) (string, error) {
	var result string
	err := util.RetryWithBackoff(5, 2*time.Second, func() error {
		request := models.JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "eth_blockNumber",
			Params:  []interface{}{},
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
			"result":   response.Result,
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
	err := util.RetryWithBackoff(5, 2*time.Second, func() error {
		requests := make([]models.JSONRPCRequest, len(blockNumbers))
		for i, blockNumber := range blockNumbers {
			requests[i] = models.JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "eth_getBlockByNumber",
				Params:  []interface{}{blockNumber, fullTx},
				ID:      int64(i + 1),
			}
		}

		var responses []models.JSONRPCResponse
		if err := jsonrpc.SendBatchJSONRPCRequest(client, apiKey, requests, &responses); err != nil {
			return err
		}

		blocks = make([]*models.Block, len(responses))
		for i, response := range responses {
			var block models.Block
			if err := json.Unmarshal(response.Result, &block); err != nil {
				return err
			}
			blocks[i] = &block
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}
