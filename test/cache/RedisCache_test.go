package test_cache

import (
	"os"
	"testing"

	cconf "github.com/pip-services3-go/pip-services3-commons-go/config"
	rediscache "github.com/pip-services3-go/pip-services3-redis-go/cache"
	redisfixture "github.com/pip-services3-go/pip-services3-redis-go/test/fixture"
)

func TestRedisCache(t *testing.T) {
	var cache *rediscache.RedisCache
	var fixture *redisfixture.CacheFixture

	host := os.Getenv("REDIS_SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_SERVICE_PORT")
	if port == "" {
		port = "6379"
	}

	cache = rediscache.NewRedisCache()
	config := cconf.NewConfigParamsFromTuples(
		"connection.host", host,
		"connection.port", port,
	)
	cache.Configure(config)
	fixture = redisfixture.NewCacheFixture(cache)
	cache.Open("")
	defer cache.Close("")

	t.Run("TestRedisCache:Store and Retrieve", fixture.TestStoreAndRetrieve)
	t.Run("TestRedisCache:Retrieve Expired", fixture.TestRetrieveExpired)
	t.Run("TestRedisCache:Remove", fixture.TestRemove)
}
