package test_fixture

import (
	"testing"

	clock "github.com/pip-services3-go/pip-services3-components-go/lock"
	"github.com/stretchr/testify/assert"
)

const (
	LOCK1 string = "lock_1"
	LOCK2 string = "lock_2"
	LOCK3 string = "lock_3"
)

type LockFixture struct {
	lock clock.ILock
}

func NewLockFixture(lock clock.ILock) *LockFixture {
	c := LockFixture{}
	c.lock = lock
	return &c
}

func (c *LockFixture) TestTryAcquireLock(t *testing.T) {

	// Try to acquire lock for the first time
	result, err := c.lock.TryAcquireLock("", LOCK1, 3000)
	assert.Nil(t, err)
	assert.True(t, result)

	// Try to acquire lock for the second time
	result, err = c.lock.TryAcquireLock("", LOCK1, 3000)
	assert.Nil(t, err)
	assert.False(t, result)

	// Release the lock
	err = c.lock.ReleaseLock("", LOCK1)

	// Try to acquire lock for the third time
	result, err = c.lock.TryAcquireLock("", LOCK1, 3000)
	assert.Nil(t, err)
	assert.True(t, result)

	c.lock.ReleaseLock("", LOCK1)
}

func (c *LockFixture) TestAcquireLock(t *testing.T) {

	// Acquire lock for the first time
	err := c.lock.AcquireLock("", LOCK2, 3000, 1000)
	assert.Nil(t, err)

	// Acquire lock for the second time
	err = c.lock.AcquireLock("", LOCK2, 3000, 1000)
	assert.NotNil(t, err)

	// Release the lock
	err = c.lock.ReleaseLock("", LOCK2)

	// Acquire lock for the third time
	err = c.lock.AcquireLock("", LOCK2, 3000, 1000)
	assert.Nil(t, err)

	c.lock.ReleaseLock("", LOCK2)
}

func (c *LockFixture) TestReleaseLock(t *testing.T) {

	// Acquire lock for the first time
	result, err := c.lock.TryAcquireLock("", LOCK3, 3000)
	assert.Nil(t, err)
	assert.True(t, result)

	// Release the lock for the first time
	err = c.lock.ReleaseLock("", LOCK3)
	assert.Nil(t, err)
	// Release the lock for the second time
	err = c.lock.ReleaseLock("", LOCK3)
	assert.Nil(t, err)
}
