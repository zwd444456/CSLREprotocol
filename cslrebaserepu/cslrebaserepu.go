package main

import (
	"code/config"
	"code/cslrebaserepu/client"
	"code/cslrebaserepu/shard1"
	"code/cslrebaserepu/shard2"
	"code/cslrebaserepu/shard3"
	"code/cslrebaserepu/shard4"
	"code/cslrebaserepu/shard5"
	"code/cslrebaserepu/shard6"
	"code/cslrebaserepu/shard7"
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

	for i := 1; i <= config.Node; i++ {
		port := 8001 + (i - 1)
		go shard1.NewNode(i, &wg).Listen(port)
	}

	for i := config.Node + 1; i <= config.Node*2; i++ {
		port := 8001 + config.Node + (i - config.Node - 1)
		go shard2.NewNode(i, &wg).Listen(port)
	}

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
	go cslrebaserepu.ListenForFeedback()
	cslrebaserepu.ExitChan = make(chan bool) // Initialize the exit channel
	go func() {
		cslrebaserepu.StartClient()
	}()

	// Block main goroutine to keep the process running until exitChan receives a signal
	<-cslrebaserepu.ExitChan
}
