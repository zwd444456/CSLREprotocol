package main

import (
	"code/config"
	cslrebaseht "code/cslrebaseht/client"
	"code/cslrebaseht/shard1"
	"code/cslrebaseht/shard2"
	"code/cslrebaseht/shard3"
	"code/cslrebaseht/shard4"
	"code/cslrebaseht/shard5"
	"code/cslrebaseht/shard6"
	"code/cslrebaseht/shard7"
	"strconv"
	"sync"
	"time"
)

func main() {
	GlobalPayload := strconv.Itoa(len(config.GlobalPayload))
	Node := strconv.Itoa(config.Node)
	Delay := strconv.Itoa(config.Delay)
	Round := strconv.Itoa(config.Round)
	MaliciousNode := strconv.Itoa(config.MaliciousNode)
	ShardNumber := strconv.Itoa(config.ShardNumber)
	str := GlobalPayload + "T" + Node + "N" + Delay + "D" + Round + "R" + MaliciousNode + "M" + ShardNumber + "S.txt"
	config.SetupLogFile(str, 100*1024*1024)
	// 定期刷新缓冲区
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			config.FlushLogFile()
		}
	}()
	var wg sync.WaitGroup

	// Initialize nodes for Shard 1
	for i := 1; i <= config.Node; i++ {
		port := 8001 + (i - 1)
		go shard1.NewNode(i, &wg).Listen(port)
	}

	// Initialize nodes for Shard 2
	for i := config.Node + 1; i <= config.Node*2; i++ {
		port := 8001 + config.Node + (i - config.Node - 1)
		go shard2.NewNode(i, &wg).Listen(port)
	}

	// Initialize nodes for Shard 3
	for i := config.Node*2 + 1; i <= config.Node*3; i++ {
		port := 8001 + config.Node*2 + (i - config.Node*2 - 1)
		go shard3.NewNode(i, &wg).Listen(port)
	}
	if config.ShardNumber >= 4 {
		for i := config.Node*3 + 1; i <= config.Node*4; i++ {
			port := 8001 + config.Node*3 + (i - config.Node*3 - 1)
			go shard4.NewNode(i, &wg).Listen(port)
		}
	}
	if config.ShardNumber >= 5 {
		for i := config.Node*4 + 1; i <= config.Node*5; i++ {
			port := 8001 + config.Node*4 + (i - config.Node*4 - 1)
			go shard5.NewNode(i, &wg).Listen(port)
		}
	}
	if config.ShardNumber >= 6 {
		for i := config.Node*5 + 1; i <= config.Node*6; i++ {
			port := 8001 + config.Node*5 + (i - config.Node*5 - 1)
			go shard6.NewNode(i, &wg).Listen(port)
		}
	}
	if config.ShardNumber >= 7 {
		for i := config.Node*6 + 1; i <= config.Node*7; i++ {
			port := 8001 + config.Node*6 + (i - config.Node*6 - 1)
			go shard7.NewNode(i, &wg).Listen(port)
		}
	}
	// Start client to send transactions
	go cslrebaseht.ListenForFeedback()
	cslrebaseht.ExitChan = make(chan bool) // Initialize the exit channel
	go func() {
		cslrebaseht.StartClient()
	}()

	// Block main goroutine to keep the process running until exitChan receives a signal
	<-cslrebaseht.ExitChan
}
