package utils

import (
	"time"
)

type Callback func(args ...interface{}) interface{}

func Now() int64 {
	return time.Now().UnixNano()/1000000
}
