package cache

import (
	"encoding/json"
	"errors"
	as "github.com/aerospike/aerospike-client-go"
	"time"
)

var (
	binLockName       = "_lock"
	binDataName       = "data"
	ErrExecutionLimit = errors.New("Exceeded execution limit")
)

type Client interface {
	Get(key string, result interface{}, executionLimit time.Duration, expire time.Duration, getter func()) error
}

type client struct {
	asClient  *as.Client
	namespace string
	set       string
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func NewClient(address string, port int, namespace string, set string) Client {
	asClient, err := as.NewClient(address, port)
	panicOnError(err)

	return &client{asClient: asClient, namespace: namespace, set: set}
}

func (c *client) lockKey(key *as.Key) error {
	writePolicy := as.NewWritePolicy(0, -1)
	writePolicy.RecordExistsAction = as.CREATE_ONLY

	lockBin := as.NewBin(binLockName, "true")

	return c.asClient.PutBins(writePolicy, key, lockBin)
}

func (c *client) writeData(key *as.Key, result interface{}, expire time.Duration) {
	writePolicy := as.NewWritePolicy(0, int32(expire.Seconds()))

	value, err := json.Marshal(result)
	panicOnError(err)

	lockBin := as.NewBin(binLockName, nil)
	dataBin := as.NewBin(binDataName, string(value))

	err = c.asClient.PutBins(writePolicy, key, lockBin, dataBin)
	panicOnError(err)
}

func (c *client) waitData(key *as.Key, result interface{}, executionLimit time.Duration) error {
	executionMaxTime := time.Now().Add(executionLimit)

	for {
		if time.Now().After(executionMaxTime) {
			return ErrExecutionLimit
		}

		if c.getData(key, result) {
			return nil
		}
	}
}

func (c *client) getData(key *as.Key, result interface{}) bool {
	policy := as.NewPolicy()
	policy.Timeout = 3 * time.Second
	policy.MaxRetries = 3

	record, err := c.asClient.Get(policy, key)
	panicOnError(err)

	if record != nil {
		if data, ok := record.Bins[binDataName]; ok {
			json.Unmarshal([]byte(data.(string)), result)
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (c *client) Get(key string, result interface{}, executionLimit time.Duration, expire time.Duration, getter func()) error {
	var err error

	asKey, err := as.NewKey(c.namespace, c.set, key)
	panicOnError(err)

	if c.getData(asKey, result) {
		return nil
	}

	err = c.lockKey(asKey)

	if err == nil {
		getter()
		c.writeData(asKey, result, expire)
	} else {
		err = c.waitData(asKey, result, executionLimit)
		if err != nil {
			return err
		}
	}

	return nil
}
