package build

import (
	cref "github.com/pip-services3-gox/pip-services3-commons-gox/refer"
	cbuild "github.com/pip-services3-gox/pip-services3-components-gox/build"
	rediscache "github.com/pip-services3-gox/pip-services3-redis-gox/cache"
	redislock "github.com/pip-services3-gox/pip-services3-redis-gox/lock"
)

/*
DefaultRedisFactory are creates Redis components by their descriptors.

See RedisCache
See RedisLock
*/
type DefaultRedisFactory struct {
	*cbuild.Factory
	Descriptor           *cref.Descriptor
	RedisCacheDescriptor *cref.Descriptor
	RedisLockDescriptor  *cref.Descriptor
}

// NewDefaultRedisFactory method are create a new instance of the factory.
func NewDefaultRedisFactory() *DefaultRedisFactory {

	c := DefaultRedisFactory{}
	c.Factory = cbuild.NewFactory()
	c.Descriptor = cref.NewDescriptor("pip-services", "factory", "redis", "default", "1.0")
	c.RedisCacheDescriptor = cref.NewDescriptor("pip-services", "cache", "redis", "*", "1.0")
	c.RedisLockDescriptor = cref.NewDescriptor("pip-services", "lock", "redis", "*", "1.0")
	c.RegisterType(c.RedisCacheDescriptor, rediscache.NewRedisCache[any])
	c.RegisterType(c.RedisLockDescriptor, redislock.NewRedisLock)
	return &c
}
