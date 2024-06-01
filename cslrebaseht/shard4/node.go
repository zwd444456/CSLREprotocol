package shard4

import (
	"code/config"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	ID                int
	IsLeader          bool
	mu                sync.Mutex
	wg                *sync.WaitGroup
	receivedResults   map[int]map[int]int  // transactionID -> nodeID -> result
	receivedAcks      map[int]map[int]bool // transactionID -> nodeID -> ack
	consensusRequired int                  // number of nodes required for consensus
	prepareCerts      map[int]int
	voteCerts         map[int]int
	completed         map[int]bool // track completed transactions
}

func NewNode(id int, wg *sync.WaitGroup) *Node {
	return &Node{
		ID:                id,
		IsLeader:          id%config.Node == 1, // 动态决定 Leader 节点
		wg:                wg,
		receivedResults:   make(map[int]map[int]int),
		receivedAcks:      make(map[int]map[int]bool),
		consensusRequired: config.Node/2 + 1, // 动态决定需要的共识数量
		prepareCerts:      make(map[int]int),
		voteCerts:         make(map[int]int),
		completed:         make(map[int]bool),
	}
}

func (n *Node) ReceiveTransaction(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("id")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	certificate := n.ProcessTransaction(transactionID)
	n.mu.Lock()
	n.wg.Add(1) // Add for each new transaction
	n.mu.Unlock()
	log.Printf("time=\"%s\" level=info msg=\"Node %d received transaction %d from client\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	n.SendResultToLeader(transactionID, certificate)

}

func (n *Node) ProcessTransaction(transactionID int) int {
	return rand.Intn(1000) // Simulated certificate
}

func (n *Node) SendResultToLeader(transactionID, result int) {
	if n.IsLeader {
		n.mu.Lock()
		defer n.mu.Unlock()
		if _, exists := n.receivedResults[transactionID]; !exists {
			n.receivedResults[transactionID] = make(map[int]int)
		}
		n.receivedResults[transactionID][n.ID] = result

		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received its own result %d for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, result, transactionID)

		n.CheckConsensus(transactionID)
	} else {
		log.Printf("time=\"%s\" level=info msg=\"Node %d sending result %d for transaction %d to leader\"\n", time.Now().Format(time.RFC3339), n.ID, result, transactionID)
		go func() {
			_, err := http.Get(fmt.Sprintf("http://localhost:%d/result?transactionID=%d&nodeID=%d&result=%d&payload=%s", config.Shard4Leader, transactionID, n.ID, result, config.GlobalPayload))
			if err != nil {
				log.Printf("time=\"%s\" level=error msg=\"Error sending result to leader: %v\"\n", time.Now().Format(time.RFC3339), err)
			}
		}()
	}
}

func (n *Node) ReceiveResult(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	nodeIDStr := r.URL.Query().Get("nodeID")
	resultStr := r.URL.Query().Get("result")

	transactionID, _ := strconv.Atoi(transactionIDStr)
	nodeID, _ := strconv.Atoi(nodeIDStr)
	result, _ := strconv.Atoi(resultStr)

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.completed[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d ignored result %d for transaction %d from Node %d because consensus is already reached\"\n", time.Now().Format(time.RFC3339), n.ID, result, transactionID, nodeID)
		return
	}

	if _, exists := n.receivedResults[transactionID]; !exists {
		n.receivedResults[transactionID] = make(map[int]int)
	}
	n.receivedResults[transactionID][nodeID] = result

	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received result %d from Node %d for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, result, nodeID, transactionID)

	n.CheckConsensus(transactionID)
}

func (n *Node) CheckConsensus(transactionID int) {
	if len(n.receivedResults[transactionID]) >= n.consensusRequired {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d reached consensus for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		n.SendPrepare(transactionID)
	}
}

func (n *Node) SendPrepare(transactionID int) {
	certificate := n.receivedResults[transactionID][n.ID]
	n.prepareCerts[transactionID] = certificate
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d sending prepare certificate %d for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, certificate, transactionID)

	for i := config.Shard4Leader; i <= config.Shard5Leader-1; i++ { // Send to other nodes
		go func(i int) {
			_, err := http.Get(fmt.Sprintf("http://localhost:%d/prepare?transactionID=%d&certificate=%d&payload=%s", i, transactionID, certificate, config.GlobalPayload))
			if err != nil {
				log.Printf("time=\"%s\" level=error msg=\"Error sending prepare certificate to Node %d: %v\"\n", time.Now().Format(time.RFC3339), i, err)
			}
		}(i)
	}
}

func (n *Node) ReceivePrepare(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	certificateStr := r.URL.Query().Get("certificate")
	time.Sleep(time.Millisecond * time.Duration(config.Delay))
	transactionID, _ := strconv.Atoi(transactionIDStr)
	certificate, _ := strconv.Atoi(certificateStr)

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.completed[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Node %d ignored prepare certificate %d for transaction %d because consensus is already reached\"\n", time.Now().Format(time.RFC3339), n.ID, certificate, transactionID)
		return
	}

	log.Printf("time=\"%s\" level=info msg=\"Node %d received prepare certificate %d for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, certificate, transactionID)

	go func() {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/prepare_ack?transactionID=%d&nodeID=%d&payload=%s", config.Shard4Leader, transactionID, n.ID, config.GlobalPayload))
		if err != nil {
			log.Printf("time=\"%s\" level=error msg=\"Error sending prepare ack to leader: %v\"\n", time.Now().Format(time.RFC3339), err)
		}
	}()
}

func (n *Node) ReceivePrepareAck(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	nodeIDStr := r.URL.Query().Get("nodeID")

	transactionID, _ := strconv.Atoi(transactionIDStr)
	nodeID, _ := strconv.Atoi(nodeIDStr)

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.completed[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d ignored prepare ack for transaction %d from Node %d because consensus is already reached\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID, nodeID)
		return
	}

	if _, exists := n.receivedAcks[transactionID]; !exists {
		n.receivedAcks[transactionID] = make(map[int]bool)
	}
	n.receivedAcks[transactionID][nodeID] = true

	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received prepare ack from Node %d for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, nodeID, transactionID)

	n.CheckPrepareAcks(transactionID)
}

func (n *Node) CheckPrepareAcks(transactionID int) {
	if len(n.receivedAcks[transactionID]) >= n.consensusRequired {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d reached prepare consensus for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		n.SendVote(transactionID)
	}
}

func (n *Node) SendVote(transactionID int) {
	certificate := n.receivedResults[transactionID][n.ID]
	n.voteCerts[transactionID] = certificate
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d sending vote certificate %d for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, certificate, transactionID)

	for i := config.Shard4Leader; i <= config.Shard5Leader-1; i++ { // Send to other nodes
		go func(i int) {
			_, err := http.Get(fmt.Sprintf("http://localhost:%d/vote?transactionID=%d&certificate=%d&payload=%s", i, transactionID, certificate, config.GlobalPayload))
			if err != nil {
				log.Printf("time=\"%s\" level=error msg=\"Error sending vote certificate to Node %d: %v\"\n", time.Now().Format(time.RFC3339), i, err)
			}
		}(i)
	}
}

func (n *Node) ReceiveVote(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	certificateStr := r.URL.Query().Get("certificate")
	time.Sleep(time.Millisecond * time.Duration(config.Delay))
	transactionID, _ := strconv.Atoi(transactionIDStr)
	certificate, _ := strconv.Atoi(certificateStr)

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.completed[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Node %d ignored vote certificate %d for transaction %d because consensus is already reached\"\n", time.Now().Format(time.RFC3339), n.ID, certificate, transactionID)
		return
	}

	log.Printf("time=\"%s\" level=info msg=\"Node %d received vote certificate %d for transaction %d \"\n", time.Now().Format(time.RFC3339), n.ID, certificate, transactionID)

	go func() {
		log.Printf("time=\"%s\" level=info msg=\"Node %d send vote ack certificate %d for transaction %d to leader %d \"\n", time.Now().Format(time.RFC3339), n.ID, certificate, transactionID, config.Shard4Leader)
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/vote_ack?transactionID=%d&nodeID=%d&payload=%s", config.Shard4Leader, transactionID, n.ID, config.GlobalPayload))
		if err != nil {
			log.Printf("time=\"%s\" level=error msg=\"Error sending vote ack to leader: %v\"\n", time.Now().Format(time.RFC3339), err)
		}
	}()
}

func (n *Node) ReceiveVoteAck(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	nodeIDStr := r.URL.Query().Get("nodeID")

	transactionID, _ := strconv.Atoi(transactionIDStr)
	nodeID, _ := strconv.Atoi(nodeIDStr)

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.completed[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d ignored vote ack for transaction %d from Node %d because consensus is already reached\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID, nodeID)
		return
	}

	if _, exists := n.receivedAcks[transactionID]; !exists {
		n.receivedAcks[transactionID] = make(map[int]bool)
	}
	n.receivedAcks[transactionID][nodeID] = true

	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received vote ack from Node %d for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, nodeID, transactionID)

	n.CheckVoteAcks(transactionID)
}

func (n *Node) CheckVoteAcks(transactionID int) {
	if len(n.receivedAcks[transactionID]) >= n.consensusRequired {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d reached vote consensus for transaction %d\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		n.SendCertificateToShard3(transactionID)
	}
}

func (n *Node) SendCertificateToShard3(transactionID int) {
	prepareCert := n.prepareCerts[transactionID]
	voteCert := n.voteCerts[transactionID]
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d sending prepare certificate %d and vote certificate %d for transaction %d to Shard 3\"\n", time.Now().Format(time.RFC3339), n.ID, prepareCert, voteCert, transactionID)

	_, err := http.Get(fmt.Sprintf("http://localhost:%d/certificate?transactionID=%d&prepareCert=%d&voteCert=%d&shard=%d&payload=%s", config.Shard3Leader, transactionID, prepareCert, voteCert, n.ID, config.GlobalPayload))
	if err != nil {
		log.Printf("time=\"%s\" level=error msg=\"Error sending certificates to Shard 3: %v\"\n", time.Now().Format(time.RFC3339), err)
	} else {
		n.completed[transactionID] = true // Mark transaction as completed
	}
}

func (n *Node) ReceiveFeedback(w http.ResponseWriter, r *http.Request) {
	transactionID := r.URL.Query().Get("id")
	log.Printf("time=\"%s\" level=info msg=\"Node %d received feedback for transaction %s\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	n.wg.Done() // Done for each completed transaction
}

func (n *Node) Listen(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/transaction", n.ReceiveTransaction)
	mux.HandleFunc("/result", n.ReceiveResult)
	mux.HandleFunc("/prepare", n.ReceivePrepare)
	mux.HandleFunc("/prepare_ack", n.ReceivePrepareAck)
	mux.HandleFunc("/vote", n.ReceiveVote)
	mux.HandleFunc("/vote_ack", n.ReceiveVoteAck)
	mux.HandleFunc("/feedback", n.ReceiveFeedback)
	log.Printf("time=\"%s\" level=info msg=\"Node %d listening on port %d\"\n", time.Now().Format(time.RFC3339), n.ID, port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
