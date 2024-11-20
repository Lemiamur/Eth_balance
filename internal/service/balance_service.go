package service

import (
	"eth_bal/configs"
	"eth_bal/internal/cache"
	"eth_bal/internal/models"
	"eth_bal/internal/usecase/webapi"
	"eth_bal/internal/util"
	"eth_bal/pkg/log"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

func EthChecker(cfg *configs.Config) models.ResultBlock {
	runtime.GOMAXPROCS(runtime.NumCPU())
	apiKey := getAPIKey(cfg)
	blockCache := cache.GetGlobalBlockCache(cfg.CacheSize)
	transactionsSet := loadCache(blockCache)
	client := createHTTPClient(cfg)
	latestBlockNumber := getLatestBlockNumber(client, apiKey)
	startBlockNumber := calculateStartBlockNumber(latestBlockNumber, cfg)
	analyzeBlocks(client, apiKey, blockCache, transactionsSet, latestBlockNumber, startBlockNumber, cfg)
	maxAddress, maxChange, sign := findMaxChangeAddress(transactionsSet)
	logMaxChangeAddress(maxAddress, maxChange)
	return models.ResultBlock{
		Address:   maxAddress,
		ChangeEth: util.WeiToEth(maxChange),
		Sign:      sign,
	}
}

func getAPIKey(cfg *configs.Config) string {
	apiKey := cfg.GETBLOCK_API_KEY
	if apiKey == "" {
		apiKey = os.Getenv("GETBLOCK_API_KEY")
	}
	if apiKey == "" {
		log.Logger.Fatal("Переменная окружения GETBLOCK_API_KEY не установлена")
	}
	return apiKey
}

func loadCache(blockCache *cache.BlockCache) *sync.Map {
	transactionsSet := &sync.Map{}
	log.Logger.WithField("cache_size", blockCache.Size()).Info("Кэш успешно загружен.")
	for _, key := range blockCache.Keys() {
		if block, found := blockCache.Get(key); found {
			for _, tx := range block.Transactions {
				transactionsSet.Store(tx.Hash, models.TransactionData{
					From:  tx.From,
					To:    tx.To,
					Value: tx.Value,
				})
			}
		}
	}
	return transactionsSet
}

func createHTTPClient(cfg *configs.Config) *http.Client {
	return &http.Client{
		Timeout: cfg.HTTPClientTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.IdleConnTimeout,
		},
	}
}

func getLatestBlockNumber(client *http.Client, apiKey string) int64 {
	latestBlockNumberHex, err := webapi.GetLatestBlockNumber(client, apiKey)
	if err != nil {
		log.Logger.Fatalf("Не удалось получить последний номер блока: %v", err)
	}
	return util.HexToInt(latestBlockNumberHex)
}

func calculateStartBlockNumber(latestBlockNumber int64, cfg *configs.Config) int64 {
	startBlockNumber := latestBlockNumber - cfg.BlocksToAnalyze
	if startBlockNumber < 0 {
		log.Logger.Fatal("Неверное значение начального номера блока")
	}
	return startBlockNumber
}

func analyzeBlocks(client *http.Client, apiKey string, blockCache *cache.BlockCache, transactionsSet *sync.Map, latestBlockNumber, startBlockNumber int64, cfg *configs.Config) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU()*2)
	for i := latestBlockNumber; i > startBlockNumber; i -= cfg.BatchSize {
		wg.Add(1)
		sem <- struct{}{}
		var batchBlocks []string
		for j := i; j > i-cfg.BatchSize && j >= startBlockNumber; j-- {
			blockNumberHex := util.IntToHex(j)
			if _, found := blockCache.Get(blockNumberHex); !found {
				batchBlocks = append(batchBlocks, blockNumberHex)
			}
		}
		if len(batchBlocks) > 0 {
			go func(batchBlocks []string) {
				defer wg.Done()
				defer func() { <-sem }()
				blocks, err := webapi.GetBlocksByNumbers(client, apiKey, batchBlocks, true)
				if err != nil {
					log.Logger.Warn("Не удалось загрузить блоки")
					return
				}
				for _, block := range blocks {
					blockCache.Add(block.Number, block)
					for _, tx := range block.Transactions {
						transactionsSet.Store(tx.Hash, models.TransactionData{
							From:  tx.From,
							To:    tx.To,
							Value: tx.Value,
						})
					}
				}
			}(batchBlocks)
		} else {
			wg.Done()
			<-sem
		}
	}
	wg.Wait()
}

func findMaxChangeAddress(transactionsSet *sync.Map) (string, *big.Int, string) {
	var allTransactions []models.TransactionData
	transactionsSet.Range(func(key, value interface{}) bool {
		transaction := value.(models.TransactionData)
		allTransactions = append(allTransactions, transaction)
		return true
	})
	fmt.Println("Всего транзакций:", len(allTransactions))
	var maxAddress string
	var sign string = "increase"
	maxChange := big.NewInt(0)
	for _, tx := range allTransactions {
		change := util.HexToBigInt(util.TrimQuotes(tx.Value))
		if change.Cmp(big.NewInt(0)) < 0 {
			sign = "decrease"
		}
		absChange := new(big.Int).Abs(change)
		if absChange.Cmp(maxChange) > 0 {
			maxChange = absChange
			maxAddress = tx.From
		}
	}
	return maxAddress, maxChange, sign
}

func logMaxChangeAddress(maxAddress string, maxChange *big.Int) {
	log.Logger.WithFields(logrus.Fields{
		"max_address": maxAddress,
		"max_change":  maxChange.String(),
	}).Info("Адрес с максимальным изменением баланса найден")
	fmt.Println("")
	fmt.Printf("Адрес с максимальным изменением баланса: %s\n", maxAddress)
	fmt.Printf("Изменение баланса: %s WEL, %s Eth\n", maxChange.String(), util.WeiToEth(maxChange).String())
}
