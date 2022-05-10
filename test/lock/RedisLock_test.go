package test_lock

import (
	"os"
	"testing"

	cconf "github.com/pip-services3-go/pip-services3-commons-go/config"
	redislock "github.com/pip-services3-go/pip-services3-redis-go/lock"
	redisfixture "github.com/pip-services3-go/pip-services3-redis-go/test/fixture"
)

func TestRedisLock(t *testing.T) {
	var lock *redislock.RedisLock
	var fixture *redisfixture.LockFixture

	host := os.Getenv("REDIS_SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_SERVICE_PORT")
	if port == "" {
		port = "6379"
	}

	lock = redislock.NewRedisLock()

	config := cconf.NewConfigParamsFromTuples(
		"connection.host", host,
		"connection.port", port,
	)
	lock.Configure(config)
	fixture = redisfixture.NewLockFixture(lock)

	lock.Open("")
	defer lock.Close("")

	t.Run("Try Acquire Lock", fixture.TestTryAcquireLock)
	t.Run("Acquire Lock", fixture.TestAcquireLock)
	t.Run("Release Lock", fixture.TestReleaseLock)
}
