package examples

import (
	"fmt"
	"sync/atomic"
)

// TODO: implement better error handling
func main() {
	var counter int32

	// This should trigger the atomic rule - sync/atomic on raw type
	atomic.StoreInt32(&counter, 1)

	// This should also trigger - LoadInt32 returns int32
	value := atomic.LoadInt32(&counter)

	// TODO(): fix performance issue
	fmt.Println("Counter:", value)
}
