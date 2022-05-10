package lock

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	cconf "github.com/pip-services3-go/pip-services3-commons-go/config"
	cdata "github.com/pip-services3-go/pip-services3-commons-go/data"
	cerr "github.com/pip-services3-go/pip-services3-commons-go/errors"
	cref "github.com/pip-services3-go/pip-services3-commons-go/refer"
	cauth "github.com/pip-services3-go/pip-services3-components-go/auth"
	ccon "github.com/pip-services3-go/pip-services3-components-go/connect"
	clock "github.com/pip-services3-go/pip-services3-components-go/lock"
)

/*
RedisLock are distributed lock that is implemented based on Redis in-memory database.

Configuration parameters:

  - connection(s):
    - discovery_key:         (optional) a key to retrieve the connection from IDiscovery
    - host:                  host name or IP address
    - port:                  port number
    - uri:                   resource URI or connection string with all parameters in it
  - credential(s):
    - store_key:             key to retrieve parameters from credential store
    - username:              user name (currently is not used)
    - password:              user password
  - options:
    - retrytimeout:         timeout in milliseconds to retry lock acquisition. (Default: 100)
    - retries:               number of retries (default: 3)
    - db_num:                database number in Redis  (default 0)

References:

- *:discovery:*:*:1.0        (optional) IDiscovery services to resolve connection
- *:credential-store:*:*:1.0 (optional) Credential stores to resolve credential

Example:

    lock = NewRedisRedis();
    lock.Configure(cconf.NewConfigParamsFromTuples(
      "host", "localhost",
      "port", 6379,
    ));

    err = lock.Open("123")
      ...

    result, err := lock.TryAcquireLock("123", "key1", 3000)
    if result {
    	// Processing...
    }
    err = lock.ReleaseLock("123", "key1")
    // Continue...
*/
type RedisLock struct {
	*clock.Lock
	connectionResolver *ccon.ConnectionResolver
	credentialResolver *cauth.CredentialResolver

	lockId  string
	timeout int
	//retries int
	dbNum int

	client redis.Conn
}

// NewRedisLock method are creates a new instance of this lock.
func NewRedisLock() *RedisLock {
	c := &RedisLock{
		connectionResolver: ccon.NewEmptyConnectionResolver(),
		credentialResolver: cauth.NewEmptyCredentialResolver(),
		lockId:             cdata.IdGenerator.NextLong(),
		timeout:            30000,
		//retries : 3,
		dbNum:  0,
		client: nil,
	}
	c.Lock = clock.InheritLock(c)
	return c
}

// Configure method are configures component by passing configuration parameters.
// Parameters:
//   - config    configuration parameters to be set.
func (c *RedisLock) Configure(config *cconf.ConfigParams) {
	c.connectionResolver.Configure(config)
	c.credentialResolver.Configure(config)

	c.timeout = config.GetAsIntegerWithDefault("options.timeout", c.timeout)
	//c.retries = config.GetAsIntegerWithDefault("options.retries", c.retries)
	c.dbNum = config.GetAsIntegerWithDefault("options.db_num", c.dbNum)
	if c.dbNum > 15 || c.dbNum < 0 {
		c.dbNum = 0
	}
}

// SetReferences method are sets references to dependent components.
// Parameters:
//   - references 	references to locate the component dependencies.
func (c *RedisLock) SetReferences(references cref.IReferences) {
	c.connectionResolver.SetReferences(references)
	c.credentialResolver.SetReferences(references)
}

// IsOpen method are checks if the component is opened.
// Returns true if the component has been opened and false otherwise.
func (c *RedisLock) IsOpen() bool {
	return c.client != nil
}

// Open method are opens the component.
// Parameters:
// 	- correlationId 	(optional) transaction id to trace execution through call chain.
// Returns: error or nil no errors occured.
func (c *RedisLock) Open(correlationId string) error {
	var connection *ccon.ConnectionParams
	var credential *cauth.CredentialParams

	connection, err := c.connectionResolver.Resolve(correlationId)

	if err == nil && connection == nil {
		err = cerr.NewConfigError(correlationId, "NO_CONNECTION", "Connection is not configured")
		return err
	}

	credential, err = c.credentialResolver.Lookup(correlationId)
	if err != nil {
		return err
	}

	var url, host, port, password string
	var dialOpts []redis.DialOption = make([]redis.DialOption, 0)

	dialOpts = append(dialOpts, redis.DialConnectTimeout(time.Duration(c.timeout)*time.Millisecond))
	dialOpts = append(dialOpts, redis.DialDatabase(c.dbNum))

	if credential != nil {
		password = credential.Password()
		dialOpts = append(dialOpts, redis.DialPassword(password))
	}

	if connection.Uri() != "" {
		url = connection.Uri()
		c.client, err = redis.DialURL(url, dialOpts...)
	} else {
		host = connection.Host()
		if host == "" {
			host = "localhost"
		}
		port = strconv.FormatInt(int64(connection.Port()), 10)
		if port == "0" {
			port = "6379"
		}
		url = host + ":" + port
		c.client, err = redis.Dial("tcp", url, dialOpts...)
	}
	return err
}

// Close method are closes component and frees used resources.
// Parameters:
//  - correlationId 	(optional) transaction id to trace execution through call chain.
// Retruns: error or nil no errors occured.
func (c *RedisLock) Close(correlationId string) error {
	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisLock) checkOpened(correlationId string) (state bool, err error) {
	if !c.IsOpen() {
		err = cerr.NewInvalidStateError(correlationId, "NOT_OPENED", "Connection is not opened")
		return false, err
	}

	return true, nil
}

// TryAcquireLock method are makes a single attempt to acquire a lock by its key.
// It returns immediately a positive or negative result.
// Parameters:
//  - correlationId     (optional) transaction id to trace execution through call chain.
//  - key               a unique lock key to acquire.
//  - ttl               a lock timeout (time to live) in milliseconds.
// Returns: a lock result or error.
func (c *RedisLock) TryAcquireLock(correlationId string, key string, ttl int64) (result bool, err error) {
	state, err := c.checkOpened(correlationId)
	if !state {
		return false, err
	}

	res, err := redis.String(c.client.Do("SET", key, c.lockId, "NX", "PX", ttl))
	if err != nil && err == redis.ErrNil {
		return false, nil
	}
	return res == "OK", err
}

// ReleaseLock method are releases prevously acquired lock by its key.
//  - correlationId     (optional) transaction id to trace execution through call chain.
//  - key               a unique lock key to release.
// Returns: error or nil for success.
func (c *RedisLock) ReleaseLock(correlationId string, key string) error {
	state, err := c.checkOpened(correlationId)
	if !state {
		return err
	}

	// Start transaction on key
	_, err = c.client.Do("WATCH", key)
	if err != nil {
		return err
	}

	// Read and check if lock is the same
	keyId, err := redis.String(c.client.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		c.client.Do("UNWATCH")
		return err
	}
	// Remove the lock if it matches
	if keyId == c.lockId {
		c.client.Send("MULTI")
		c.client.Send("DEL", key)
		_, err = c.client.Do("EXEC")
	} else { // Cancel transaction if it doesn"t match
		_, err = c.client.Do("UNWATCH")
	}
	return err
}
