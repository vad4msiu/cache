package main

import (
	"cache/client"
	. "cache/shared"
	"flag"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	host          = flag.String("h", "127.0.0.1", "Aerospike hostname")
	port          = flag.Int("p", 3000, "Aerospike port")
	workersCount  = flag.Int("c", 10, "Workers count")
	numOperations = flag.Int("n", 1000, "Number of operations")
)

var client cache.Client

func runRequests(num int) {
	for i := 0; i < num; i++ {
		result := &Result{}

		getter := func() {
			result.Data = rand.Intn(1000000)
		}
		key := strconv.Itoa(rand.Intn(1000000))

		err := client.Get(key, result, ExecutionLimit, Expire, getter)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	var wg sync.WaitGroup

	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	client = cache.NewClient(*host, *port, Namespace, Set)
	numOperationsPerWorker := *numOperations / *workersCount

	for i := 0; i < *workersCount; i++ {
		wg.Add(1)
		go func() {
			runRequests(numOperationsPerWorker)
			wg.Done()
		}()
	}

	wg.Wait()
}
