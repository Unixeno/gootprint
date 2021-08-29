package sdk

import (
	"fmt"
	"sync/atomic"
)

var counter uint64

func Collect(info string) {
	atomic.AddUint64(&counter, 1)
	fmt.Println(info)
}

func ResetCounter() {
	atomic.StoreUint64(&counter, 0)
}

func GetCounter() uint64 {
	return atomic.LoadUint64(&counter)
}
