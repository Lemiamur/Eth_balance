package cache

import (
	"encoding/gob"
	"os"

	models "eth_bal/internal/models"
	"eth_bal/pkg/log"

	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
)

type BlockCache struct {
	cache *lru.Cache
}

// Инициализация LRU-кэша с максимальным размером
func NewBlockCache(size int) (*BlockCache, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &BlockCache{cache: cache}, nil
}

// Получение блока из кэша
func (c *BlockCache) Get(blockNumber string) (*models.Block, bool) {
	block, ok := c.cache.Get(blockNumber)
	if ok {
		log.Logger.WithFields(logrus.Fields{
			"block_number": blockNumber,
		}).Debug("Cache hit")
	} else {
		log.Logger.WithFields(logrus.Fields{
			"block_number": blockNumber,
		}).Debug("Cache miss")
		return nil, false
	}

	if block == nil {
		log.Logger.WithFields(logrus.Fields{
			"block_number": blockNumber,
		}).Debug("Cache value is nil")
		return nil, false
	}

	return block.(*models.Block), true
}

// Добавление блока в кэш
func (c *BlockCache) Add(blockNumber string, block *models.Block) {
	c.cache.Add(blockNumber, block)
	log.Logger.WithFields(logrus.Fields{
		"block_number": blockNumber,
	}).Debug("Block added to cache")
}

// Сохранение кэша на диск
func (c *BlockCache) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	cacheMap := make(map[string]*models.Block)
	for _, key := range c.cache.Keys() {
		if value, ok := c.cache.Get(key); ok {
			cacheMap[key.(string)] = value.(*models.Block)
		}
	}
	return encoder.Encode(cacheMap)
}

// Загрузка кэша с диска
func (c *BlockCache) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	cacheMap := make(map[string]*models.Block)
	if err := decoder.Decode(&cacheMap); err != nil {
		return err
	}

	for key, value := range cacheMap {
		c.cache.Add(key, value)
	}
	return nil
}

// Получение размера кэша
func (c *BlockCache) Size() int {
	return c.cache.Len()
}

// Получение всех ключей из кэша
func (c *BlockCache) Keys() []interface{} {
	return c.cache.Keys()
}
