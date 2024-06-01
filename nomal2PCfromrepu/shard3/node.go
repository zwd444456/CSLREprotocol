package shard3

import (
	"code/config"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	ID                    int
	IsLeader              bool
	NumCertificatesNeeded int
	Certificates          map[int]map[int]bool
	ReceivedCertificates  map[int]int
	ShardCertificates     map[int]map[int]bool // 记录每个交易的分片证书接收情况
	ConfirmedTransactions map[int]bool         // 记录每个交易是否已经确认
	consensusVotes        map[int]map[int]bool // 记录每个交易的共识投票情况
	consensusRequired     int
	mu                    sync.Mutex
	wg                    *sync.WaitGroup
}

func NewNode(id int, wg *sync.WaitGroup) *Node {
	return &Node{
		ID:                    id,
		IsLeader:              id%config.Node == 1,
		NumCertificatesNeeded: config.ShardNumber - 1, // We need certificates from two different shards
		Certificates:          make(map[int]map[int]bool),
		ReceivedCertificates:  make(map[int]int),
		ShardCertificates:     make(map[int]map[int]bool),
		ConfirmedTransactions: make(map[int]bool),
		consensusRequired:     config.Node / 2, // 动态决定需要的共识数量
		consensusVotes:        make(map[int]map[int]bool),
		wg:                    wg,
	}
}

func (n *Node) ReceiveTransaction(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("id")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	n.mu.Lock()
	if _, exists := n.Certificates[transactionID]; !exists {
		n.Certificates[transactionID] = make(map[int]bool)
		n.ReceivedCertificates[transactionID] = 0
		n.ShardCertificates[transactionID] = make(map[int]bool)
	}
	n.mu.Unlock()
	log.Printf("time=\"%s\" level=info msg=\"Node %d received transaction %d from client.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
}

func (n *Node) ReceiveCertificate(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	shardIDStr := r.URL.Query().Get("shard")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	shardID, _ := strconv.Atoi(shardIDStr)

	n.mu.Lock()
	defer n.mu.Unlock()
	// 检查交易是否已经确认
	if n.ConfirmedTransactions[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Transaction %d is already confirmed, ignoring certificate from Shard %d.\"\n", time.Now().Format(time.RFC3339), transactionID, shardID)
		return
	}

	shardGroup := (shardID-1)/config.Node + 1

	// 初始化子 map
	if _, exists := n.Certificates[transactionID]; !exists {
		n.Certificates[transactionID] = make(map[int]bool)
	}
	if _, exists := n.ShardCertificates[transactionID]; !exists {
		n.ShardCertificates[transactionID] = make(map[int]bool)
	}

	if !n.ShardCertificates[transactionID][shardGroup] { // 检查当前分片组是否已处理过
		n.Certificates[transactionID][shardID] = true
		n.ReceivedCertificates[transactionID]++
		n.ShardCertificates[transactionID][shardGroup] = true
		log.Printf("time=\"%s\" level=info msg=\"Node %d received certificate for transaction %d from Shard Group %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID, shardGroup)
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received is %d.\"\n", time.Now().Format(time.RFC3339), n.ID, len(n.ShardCertificates[transactionID]))
		n.CheckCertificates(transactionID)
	} else {
		log.Printf("time=\"%s\" level=info msg=\"Node %d already received certificate for transaction %d from Shard Group %d node %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID, shardGroup, shardID)
	}
}

func (n *Node) CheckCertificates(transactionID int) {
	if n.ReceivedCertificates[transactionID] == n.NumCertificatesNeeded {
		if n.IsLeader {
			n.StartConsensus(transactionID)
		} else {
			log.Printf("time=\"%s\" level=info msg=\"Node %d received all certificates for transaction %d. Waiting for leader to start consensus.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		}
	}
}

func (n *Node) StartConsensus(transactionID int) {
	if _, exists := n.consensusVotes[transactionID]; !exists {
		n.consensusVotes[transactionID] = make(map[int]bool)
	}

	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d starting consensus for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)

	// Send consensus request to other nodes
	for i := config.Shard3Leader; i <= config.Shard3Leader+config.Node-1; i++ { // Nodes 10, 11, 12
		if i != n.ID {
			go func(i int) {
				_, err := http.Get(fmt.Sprintf("http://localhost:%d/consensus?transactionID=%d&payload=%s", i, transactionID, config.GlobalPayload))
				if err != nil {
					log.Printf("time=\"%s\" level=error msg=\"Error sending consensus request to Node %d: %v\"\n", time.Now().Format(time.RFC3339), i, err)
				}
			}(i)
		}
	}
}

func (n *Node) ReceiveVote(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	nodeIDStr := r.URL.Query().Get("nodeID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	nodeID, _ := strconv.Atoi(nodeIDStr)
	time.Sleep(time.Millisecond * time.Duration(2*config.Delay))
	time.Sleep(time.Millisecond * time.Duration(20))
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.ConfirmedTransactions[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Transaction %d is already confirmed, ignoring vote from Node %d.\"\n", time.Now().Format(time.RFC3339), transactionID, nodeID)
		return
	}

	if _, exists := n.consensusVotes[transactionID]; !exists {
		n.consensusVotes[transactionID] = make(map[int]bool)
	}
	n.consensusVotes[transactionID][nodeID] = true

	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received vote from Node %d for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, nodeID, transactionID)

	n.CheckConsensusVotes(transactionID)
}

func (n *Node) CheckConsensusVotes(transactionID int) {
	if len(n.consensusVotes[transactionID]) >= n.consensusRequired {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d reached consensus for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		n.ConfirmTransaction(transactionID)
	}
}

func (n *Node) ConfirmTransaction(transactionID int) {
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d confirming transaction %d and sending feedback.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	n.SendFeedback(transactionID)

	n.ConfirmedTransactions[transactionID] = true
}

func (n *Node) ReceiveConsensus(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	time.Sleep(time.Millisecond * time.Duration(config.Delay))
	log.Printf("time=\"%s\" level=info msg=\"Node %d received consensus request for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	go func() {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/vote?transactionID=%d&nodeID=%d&payload=%s", config.Shard3Leader, transactionID, n.ID, config.GlobalPayload))
		if err != nil {
			log.Printf("time=\"%s\" level=error msg=\"Error sending vote to leader: %v\"\n", time.Now().Format(time.RFC3339), err)
		}
	}()
}

func (n *Node) SendFeedback(transactionID int) {
	time.Sleep(time.Millisecond * time.Duration(config.Delay))
	time.Sleep(time.Millisecond * time.Duration(config.Delay))
	time.Sleep(time.Millisecond * time.Duration(config.Delay) / 40)
	if config.Delay != 100 {
		time.Sleep(time.Millisecond * time.Duration(len(config.GlobalPayload)) / 10)
	}
	log.Printf("time=\"%s\" level=info msg=\"Node %d received all certificates for transaction %d and sending feedback.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	feedbackURL1 := fmt.Sprintf("http://localhost:%d/feedback?id=%d&payload=%s", config.Shard1Leader, transactionID, config.GlobalPayload)
	feedbackURL2 := fmt.Sprintf("http://localhost:%d/feedback?id=%d&payload=%s", config.Shard2Leader, transactionID, config.GlobalPayload)
	clientFeedbackURL := fmt.Sprintf("http://localhost:8000/feedback?id=%d&payload=%s", transactionID, config.GlobalPayload)
	go http.Get(feedbackURL1)
	go http.Get(feedbackURL2)
	go http.Get(clientFeedbackURL)

	// 标记交易已确认
	n.ConfirmedTransactions[transactionID] = true

	// Clean up for the next transaction
	delete(n.Certificates, transactionID)
	delete(n.ReceivedCertificates, transactionID)
	delete(n.ShardCertificates, transactionID)
}

func (n *Node) Listen(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/transaction", n.ReceiveTransaction)
	mux.HandleFunc("/certificate", n.ReceiveCertificate)
	mux.HandleFunc("/consensus", n.ReceiveConsensus)
	mux.HandleFunc("/vote", n.ReceiveVote)
	log.Printf("time=\"%s\" level=info msg=\"Node %d listening on port %d.\"\n", time.Now().Format(time.RFC3339), n.ID, port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
