package sdk

import (
	"fmt"
	"github.com/silentred/gid"
	"sync/atomic"
)

var eventCounter uint32

func NewE(filename, path string) uint16 {
	eventID := uint16(atomic.AddUint32(&eventCounter, 1))
	fmt.Printf("register event %d for %s\n", eventID, path)
	return eventID
}

func RegisterFile(filename string) struct{} {
	fmt.Println("register file: ", filename)
	return struct{}{}
}

func C(id int64, x uint16) {
	fmt.Printf("collect event: [%d] %d\n", id, x)
}

func Call(_ uint16) int64 {
	return gid.Get()
}

func Bind(parent int64) {
	fmt.Printf("bind parernt: %d:%d\n", parent, gid.Get())
}
