package nomal2PCfromrepu

import (
	"bytes"
	"code/config"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

type Transaction struct {
	ID int
}

var (
	mu        sync.Mutex
	startTime time.Time
	results   []string
	doneChan  chan bool
	ExitChan  chan bool // New channel for exit signal
)

func sendTransaction(shardURL string, transaction Transaction, shardID, nodeID int) {
	log.Printf("time=\"%s\" level=info msg=\"Client sent transaction %d to Shard %d Node %d at %s\"\n", time.Now().Format(time.RFC3339), transaction.ID, shardID, nodeID, shardURL)

	// Create a 400-byte empty payload
	payload := bytes.Repeat([]byte{0}, 400)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/transaction?id=%d", shardURL, transaction.ID), bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("time=\"%s\" level=error msg=\"Error creating request for Shard %d Node %d at %s: %v\"\n", time.Now().Format(time.RFC3339), shardID, nodeID, shardURL, err)
		doneChan <- true
		return
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("time=\"%s\" level=error msg=\"Error sending transaction to Shard %d Node %d at %s: %v\"\n", time.Now().Format(time.RFC3339), shardID, nodeID, shardURL, err)
	}

	doneChan <- true // Signal that the transaction has been sent
}

func StartClient() {
	rand.Seed(time.Now().UnixNano())
	startTime = time.Now()
	begintime := time.Now()

	nodeURLs := []struct {
		URL    string
		Shard  int
		NodeID int
	}{}

	// Initialize node URLs based on config.Node
	for shardID := 1; shardID <= config.ShardNumber; shardID++ {
		for nodeID := 1; nodeID <= config.Node; nodeID++ {
			port := 8000 + (shardID-1)*config.Node + nodeID
			url := fmt.Sprintf("http://localhost:%d", port)
			nodeURLs = append(nodeURLs, struct {
				URL    string
				Shard  int
				NodeID int
			}{
				URL:    url,
				Shard:  shardID,
				NodeID: nodeID,
			})
		}
	}

	for i := 1; i <= config.Round; i++ {
		transaction := Transaction{ID: i}
		doneChan = make(chan bool, len(nodeURLs)) // Buffered channel to handle signals

		for _, node := range nodeURLs {
			go sendTransaction(node.URL, transaction, node.Shard, node.NodeID)
		}

		// Wait for all transactions to be sent
		for j := 0; j < len(nodeURLs); j++ {
			<-doneChan
		}

		// Wait for feedback before proceeding to the next transaction
		<-doneChan
	}

	elapsedTime := time.Since(begintime)
	log.Printf("total time:%s", elapsedTime)
	payloadLength := len(config.GlobalPayload)
	payloadLengthFloat := float64(payloadLength * config.Round)

	// Convert elapsedTime to milliseconds and then to float64
	elapsedMilliseconds := float64(elapsedTime.Milliseconds())
	rate := payloadLengthFloat / elapsedMilliseconds * 1000
	log.Printf("total TPS:%f", rate)
	config.FlushLogFile()

	SaveResults()

	// Signal main function to exit
	ExitChan <- true
}

func ReceiveFeedback(w http.ResponseWriter, r *http.Request) {
	transactionID := r.URL.Query().Get("id")
	log.Printf("time=\"%s\" level=info msg=\"Client received feedback for transaction %s\"\n", time.Now().Format(time.RFC3339), transactionID)

	elapsedTime := time.Since(startTime)
	log.Printf("paste time:%s", elapsedTime)
	startTime = time.Now()

	log.Printf("paste time:%s", elapsedTime)
	// Append the result to the results slice
	mu.Lock()
	results = append(results, fmt.Sprintf("Transaction %s: %s", transactionID, elapsedTime))
	log.Printf("paste time:%s", elapsedTime)
	mu.Unlock()
	doneChan <- true // Signal that feedback has been received
}

func ListenForFeedback() {
	http.HandleFunc("/feedback", ReceiveFeedback)
	log.Printf("time=\"%s\" level=info msg=\"Client listening for feedback on port 8000\"\n", time.Now().Format(time.RFC3339))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("time=\"%s\" level=error msg=\"Failed to start feedback listener: %v\"\n", time.Now().Format(time.RFC3339), err)
	}
}

func SaveResults() {
	file, err := os.Create("results.txt")
	if err != nil {
		log.Fatalf("time=\"%s\" level=error msg=\"Failed to create results file: %v\"\n", time.Now().Format(time.RFC3339), err)
	}
	defer file.Close()

	for _, result := range results {
		_, err := file.WriteString(result + "\n")
		if err != nil {
			log.Fatalf("time=\"%s\" level=error msg=\"Failed to write to results file: %v\"\n", time.Now().Format(time.RFC3339), err)
		}
	}
	log.Printf("time=\"%s\" level=info msg=\"Results saved to results.txt\"\n", time.Now().Format(time.RFC3339))
}
