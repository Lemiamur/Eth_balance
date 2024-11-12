package service

import (
	"eth_bal/configs"
	"eth_bal/internal/api"
	"eth_bal/internal/cache"
	"eth_bal/internal/models"
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

func EthChecker(cfg *configs.Config) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	apiKey := getAPIKey()
	blockCache := initializeCache(cfg)
	transactionsSet := loadCache(blockCache)
	client := createHTTPClient(cfg)
	latestBlockNumber := getLatestBlockNumber(client, apiKey)
	startBlockNumber := calculateStartBlockNumber(latestBlockNumber, cfg)

	analyzeBlocks(client, apiKey, blockCache, transactionsSet, latestBlockNumber, startBlockNumber, cfg)

	maxAddress, maxChange := findMaxChangeAddress(transactionsSet)
	logMaxChangeAddress(maxAddress, maxChange)

	saveCache(blockCache)
}

func getAPIKey() string {
	apiKey := os.Getenv("GETBLOCK_API_KEY")
	if apiKey == "" {
		log.Logger.Fatal("Переменная окружения GETBLOCK_API_KEY не установлена")
	}
	return apiKey
}

func initializeCache(cfg *configs.Config) *cache.BlockCache {
	blockCache, err := cache.NewBlockCache(cfg.CacheSize)
	if err != nil {
		log.Logger.Fatalf("Ошибка инициализации кэша: %v", err)
	}
	return blockCache
}

func loadCache(blockCache *cache.BlockCache) *sync.Map {
	transactionsSet := &sync.Map{}
	if err := blockCache.LoadFromFile("block_cache.gob"); err != nil {
		log.Logger.Warn("Не удалось загрузить кэш. Начинаем с пустого кэша.")
	} else {
		log.Logger.WithField("cache_size", blockCache.Size()).Info("Кэш успешно загружен.")
		for _, key := range blockCache.Keys() {
			if block, found := blockCache.Get(key.(string)); found {
				for _, tx := range block.Transactions {
					transactionsSet.Store(tx.Hash, models.TransactionData{
						From:  tx.From,
						To:    tx.To,
						Value: tx.Value,
					})
				}
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
	latestBlockNumberHex, err := api.GetLatestBlockNumber(client, apiKey)
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

				blocks, err := api.GetBlocksByNumbers(client, apiKey, batchBlocks, true)
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

func findMaxChangeAddress(transactionsSet *sync.Map) (string, *big.Int) {
	var allTransactions []models.TransactionData
	transactionsSet.Range(func(key, value interface{}) bool {
		transaction := value.(models.TransactionData)
		allTransactions = append(allTransactions, transaction)
		return true
	})

	fmt.Println("Всего транзакций:", len(allTransactions))

	var maxAddress string
	maxChange := big.NewInt(0)
	for _, tx := range allTransactions {
		absChange := new(big.Int).Abs(util.HexToBigInt(util.TrimQuotes(tx.Value)))
		if absChange.Cmp(maxChange) > 0 {
			maxChange = absChange
			maxAddress = tx.From
		}
	}
	return maxAddress, maxChange
}

func logMaxChangeAddress(maxAddress string, maxChange *big.Int) {
	log.Logger.WithFields(logrus.Fields{
		"max_address": maxAddress,
		"max_change":  maxChange.String(),
	}).Info("Адрес с максимальным изменением баланса найден")

	fmt.Println("")

	fmt.Printf("Адрес с максимальным изменением баланса: %s\n", maxAddress)
	fmt.Printf("Изменение баланса: %s WEI, %s Eth\n", maxChange.String(), util.WeiToEth(maxChange).String())
}

func saveCache(blockCache *cache.BlockCache) {
	fmt.Println("")
	if err := blockCache.SaveToFile("block_cache.gob"); err != nil {
		log.Logger.Errorf("Не удалось сохранить кэш в файл: %v", err)
	} else {
		log.Logger.Info("Кэш успешно сохранён.")
	}
}
