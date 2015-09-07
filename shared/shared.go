package cache

import (
	"time"
)

const (
	ExecutionLimit = 5 * time.Second
	Expire         = -1 * time.Second
	Namespace      = "test"
	Set            = "test"
)

type Result struct {
	Data int `json:"data"`
}

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
