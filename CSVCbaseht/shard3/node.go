package shard3

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"code/config"
)

type Node struct {
	ID                    int
	IsLeader              bool
	NumCertificatesNeeded int
	Certificates          map[int]map[int]bool
	ReceivedCertificates  map[int]int
	ShardCertificates     map[int]map[int]bool // 记录每个交易的分片证书接收情况
	ConfirmedTransactions map[int]bool         // 记录每个交易是否已经确认
	consensusRequired     int                  // number of nodes required for consensus
	consensusVotes        map[int]map[int]bool // 记录每个交易的共识投票情况
	leaderTimeouts        map[int]*time.Timer  // 记录每个交易的计时器
	mu                    sync.Mutex
	wg                    *sync.WaitGroup
	leaderChangeActive    bool // Flag to indicate leader change is in progress
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
		consensusVotes:        make(map[int]map[int]bool),
		leaderTimeouts:        make(map[int]*time.Timer),
		consensusRequired:     config.Node / 2, // 动态决定需要的共识数量
		wg:                    wg,
		leaderChangeActive:    false,
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

	if n.IsLeader {
		n.leaderTimeouts[transactionID] = time.AfterFunc(time.Millisecond*time.Duration(3*config.Delay), func() {
			n.CheckShard2Leader(transactionID)
		})
	}
}

func (n *Node) CheckShard2Leader(transactionID int) {
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d checking Shard 2 leader for transaction %d and waitting 2 个最大往返时延.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)

	go func() {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/check_leader?transactionID=%d", config.Shard2Leader, transactionID))
		if err != nil {
			log.Printf("time=\"%s\" level=error msg=\"Error checking Shard 2 leader: %v\"\n", time.Now().Format(time.RFC3339), err)
			n.InitiateLeaderChangeConsensus(transactionID)
			return
		}

		time.Sleep(time.Millisecond * time.Duration(2*config.Delay))

		n.mu.Lock()
		defer n.mu.Unlock()

		if n.ReceivedCertificates[transactionID] < n.NumCertificatesNeeded && !n.leaderChangeActive {
			log.Printf("time=\"%s\" level=info msg=\"Leader Node %d did not receive enough certificates for transaction %d, initiating leader change.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
			n.InitiateLeaderChangeConsensus(transactionID)
		} else {
			log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received enough certificates for transaction %d, no leader change needed.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		}
	}()
}

func (n *Node) ReceiveCheckLeader(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	log.Printf("time=\"%s\" level=info msg=\"Node %d received check leader request for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.leaderTimeouts[transactionID] != nil {
		n.leaderTimeouts[transactionID].Stop()
	}

	go func() {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/leader_ack?transactionID=%d&nodeID=%d", config.Shard3Leader, transactionID, n.ID))
		if err != nil {
			log.Printf("time=\"%s\" level=error msg=\"Error sending leader ack to leader: %v\"\n", time.Now().Format(time.RFC3339), err)
		}
	}()
}

func (n *Node) ReceiveLeaderAck(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	nodeIDStr := r.URL.Query().Get("nodeID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	nodeID, _ := strconv.Atoi(nodeIDStr)
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.leaderTimeouts[transactionID] != nil {
		n.leaderTimeouts[transactionID].Stop()
	}
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received leader ack from Node %d for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, nodeID, transactionID)
}

func (n *Node) InitiateLeaderChangeConsensus(transactionID int) {
	n.leaderChangeActive = true
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d initiating leader change consensus for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	if _, exists := n.consensusVotes[transactionID]; !exists {
		n.consensusVotes[transactionID] = make(map[int]bool)
	}

	// Send leader change request to other nodes
	for i := config.Shard3Leader; i <= config.Shard3Leader+config.Node-1; i++ { // Nodes 10, 11, 12
		if i != n.ID {
			go func(i int) {
				_, err := http.Get(fmt.Sprintf("http://localhost:%d/leader_change?transactionID=%d&maliciousLeaderID=%d", i, transactionID, config.Shard2Leader-8000))
				if err != nil {
					log.Printf("time=\"%s\" level=error msg=\"Error sending leader change request to Node %d: %v\"\n", time.Now().Format(time.RFC3339), i, err)
				}
			}(i)
		}
	}
}

func (n *Node) ReceiveLeaderChange(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	maliciousLeaderIDStr := r.URL.Query().Get("maliciousLeaderID")
	maliciousLeaderID, _ := strconv.Atoi(maliciousLeaderIDStr)

	log.Printf("time=\"%s\" level=info msg=\"Node %d received leader change request for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)

	go func() {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/leader_change_vote?transactionID=%d&nodeID=%d&maliciousLeaderID=%d", config.Shard3Leader, transactionID, n.ID, maliciousLeaderID))
		if err != nil {
			log.Printf("time=\"%s\" level=error msg=\"Error sending leader change vote to leader: %v\"\n", time.Now().Format(time.RFC3339), err)
		}
	}()
}

func (n *Node) ReceiveLeaderChangeVote(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	nodeIDStr := r.URL.Query().Get("nodeID")
	maliciousLeaderIDStr := r.URL.Query().Get("maliciousLeaderID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	nodeID, _ := strconv.Atoi(nodeIDStr)
	maliciousLeaderID, _ := strconv.Atoi(maliciousLeaderIDStr)

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.ConfirmedTransactions[transactionID] {
		log.Printf("time=\"%s\" level=info msg=\"Transaction %d is already confirmed, ignoring leader change vote from Node %d.\"\n", time.Now().Format(time.RFC3339), transactionID, nodeID)
		return
	}

	if _, exists := n.consensusVotes[transactionID]; !exists {
		n.consensusVotes[transactionID] = make(map[int]bool)
	}
	n.consensusVotes[transactionID][nodeID] = true

	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d received leader change vote from Node %d for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, nodeID, transactionID)

	n.CheckLeaderChangeConsensus(transactionID, maliciousLeaderID)
}

func (n *Node) CheckLeaderChangeConsensus(transactionID, maliciousLeaderID int) {
	if len(n.consensusVotes[transactionID]) >= n.consensusRequired {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d reached leader change consensus for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		n.ConfirmLeaderChange(transactionID, maliciousLeaderID)
	}
}

func (n *Node) ConfirmLeaderChange(transactionID, maliciousLeaderID int) {
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d confirming leader change for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	n.SendLeaderChangeCertificate(transactionID, maliciousLeaderID)

	n.ConfirmedTransactions[transactionID] = false
	n.leaderChangeActive = false
}

func (n *Node) SendLeaderChangeCertificate(transactionID, maliciousLeaderID int) {
	time.Sleep(time.Millisecond * time.Duration(1*config.Delay))
	for i := config.Shard2Leader; i <= config.Shard3Leader-1; i++ { // Nodes 6, 7, 8
		if i != n.ID {
			go func(i int) {
				log.Printf("time=\"%s\" level=info msg=\"Node %d sending leader change certificate for transaction %d to node %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID, i)
				_, err := http.Get(fmt.Sprintf("http://localhost:%d/leader_change_cert?id=%d&maliciousLeaderID=%d", i, transactionID, maliciousLeaderID))
				if err != nil {
					log.Printf("time=\"%s\" level=error msg=\"Error sending leader change certificate: %v\"\n", time.Now().Format(time.RFC3339), err)
				}
			}(i)
		}
	}
}

func (n *Node) ReceiveCertificate(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	shardIDStr := r.URL.Query().Get("shard")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	shardID, _ := strconv.Atoi(shardIDStr)

	n.mu.Lock()
	defer n.mu.Unlock()

	// 检查交易是否已经确认
	if n.ConfirmedTransactions[transactionID] && !n.leaderChangeActive {
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
		n.CheckCertificates(transactionID)
	} else {
		log.Printf("time=\"%s\" level=info msg=\"Node %d already received certificate for transaction %d from Shard Group %d node %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID, shardGroup, shardID)
	}
}

func (n *Node) CheckCertificates(transactionID int) {
	if n.ReceivedCertificates[transactionID] == n.NumCertificatesNeeded && !n.leaderChangeActive {
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

func (n *Node) CheckConsensusVotes(transactionID int) {
	if len(n.consensusVotes[transactionID]) >= n.consensusRequired {
		log.Printf("time=\"%s\" level=info msg=\"Leader Node %d reached consensus for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
		n.ConfirmTransaction(transactionID)
	}
}

func (n *Node) SendFeedback(transactionID int) {
	time.Sleep(time.Millisecond * time.Duration(1*config.Delay))
	log.Printf("time=\"%s\" level=info msg=\"Node %d received all certificates for transaction %d and sending feedback.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	feedbackURL1 := fmt.Sprintf("http://localhost:%d/feedback?id=%d&payload=%s", config.Shard1Leader, transactionID, config.GlobalPayload)
	feedbackURL2 := fmt.Sprintf("http://localhost:%d/feedback?id=%d&payload=%s", config.Shard2Leader, transactionID, config.GlobalPayload)
	clientFeedbackURL := fmt.Sprintf("http://localhost:8000/feedback?id=%d&payload=%s", transactionID, config.GlobalPayload)
	go http.Get(feedbackURL1)
	go http.Get(feedbackURL2)
	go http.Get(clientFeedbackURL)
	// Clean up for the next transaction
	delete(n.Certificates, transactionID)
	delete(n.ReceivedCertificates, transactionID)
	delete(n.ShardCertificates, transactionID)
}

func (n *Node) ConfirmTransaction(transactionID int) {
	log.Printf("time=\"%s\" level=info msg=\"Leader Node %d confirming transaction %d and sending feedback.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)
	n.SendFeedback(transactionID)
	n.ConfirmedTransactions[transactionID] = true
}

func (n *Node) ReceiveVote(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	nodeIDStr := r.URL.Query().Get("nodeID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	nodeID, _ := strconv.Atoi(nodeIDStr)
	time.Sleep(time.Millisecond * time.Duration(config.Delay))
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.ConfirmedTransactions[transactionID] && !n.leaderChangeActive {
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

func (n *Node) ReceiveConsensus(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transactionID")
	transactionID, _ := strconv.Atoi(transactionIDStr)
	time.Sleep(time.Millisecond * time.Duration(1*config.Delay))
	time.Sleep(time.Millisecond * time.Duration(config.Delay) / 30)
	time.Sleep(time.Millisecond * time.Duration(len(config.GlobalPayload)) / 10)
	log.Printf("time=\"%s\" level=info msg=\"Node %d received consensus request for transaction %d.\"\n", time.Now().Format(time.RFC3339), n.ID, transactionID)

	go func() {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/vote?transactionID=%d&nodeID=%d&payload=%s", config.Shard3Leader, transactionID, n.ID, config.GlobalPayload))
		if err != nil {
			log.Printf("time=\"%s\" level=error msg=\"Error sending vote to leader: %v\"\n", time.Now().Format(time.RFC3339), err)
		}
	}()
}

func (n *Node) Listen(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/transaction", n.ReceiveTransaction)
	mux.HandleFunc("/certificate", n.ReceiveCertificate)
	mux.HandleFunc("/consensus", n.ReceiveConsensus)
	mux.HandleFunc("/vote", n.ReceiveVote)
	mux.HandleFunc("/check_leader", n.ReceiveCheckLeader)
	mux.HandleFunc("/leader_ack", n.ReceiveLeaderAck)
	mux.HandleFunc("/leader_change", n.ReceiveLeaderChange)
	mux.HandleFunc("/leader_change_vote", n.ReceiveLeaderChangeVote)
	log.Printf("time=\"%s\" level=info msg=\"Node %d listening on port %d.\"\n", time.Now().Format(time.RFC3339), n.ID, port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
