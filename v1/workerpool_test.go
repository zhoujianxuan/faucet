package v1

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWorker_Submit(t *testing.T) {
	var wg sync.WaitGroup
	var workerPool = NewWorkerPool(5)

	for i := 0; i < 100; i++ {
		var id = i
		wg.Add(1)
		workerPool.Submit(func() {
			fmt.Println("id:", id)
			wg.Done()
		})
	}

	wg.Wait()
	workerPool.CloseAll()
	time.Sleep(2 * time.Second)
}
