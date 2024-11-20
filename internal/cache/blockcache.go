package cache

import (
	"eth_bal/internal/models"
	"eth_bal/pkg/log"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
)

var (
	globalBlockCache *BlockCache
	once             sync.Once
)

type BlockCache struct {
	cache *lru.Cache
}

func NewBlockCache(size int) (*BlockCache, error) {
	log.Logger.WithField("cache_size", size).Info("Initializing cache")
	cache, err := lru.New(size)
	if err != nil {
		log.Logger.WithError(err).Error("Failed to initialize cache")
		return nil, err
	}
	return &BlockCache{cache: cache}, nil
}

func GetGlobalBlockCache(size int) *BlockCache {
	once.Do(func() {
		var err error
		globalBlockCache, err = NewBlockCache(size)
		if err != nil {
			log.Logger.Fatal("Failed to initialize global block cache")
		}
	})
	return globalBlockCache
}

func (c *BlockCache) Get(blockNumber string) (*models.Block, bool) {
	I := log.Logger.WithFields(logrus.Fields{
		"block_number": blockNumber,
	})
	block, ok := c.cache.Get(blockNumber)
	if !ok {
		I.Debug("Cache miss")
		return nil, false
	}
	I.Debug("Cache hit")
	return block.(*models.Block), true
}

func (c *BlockCache) Add(blockNumber string, block *models.Block) {
	c.cache.Add(blockNumber, block)
	log.Logger.WithFields(logrus.Fields{
		"block_number": blockNumber,
	}).Debug("Block added to cache")
}

func (c *BlockCache) Size() int {
	return c.cache.Len()
}

func (c *BlockCache) Keys() []string {
	keys := c.cache.Keys()
	stringKeys := make([]string, len(keys))
	for i, key := range keys {
		stringKeys[i] = key.(string)
	}
	return stringKeys
}
