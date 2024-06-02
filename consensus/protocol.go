package consensus

import (
	"bufio"
	"context"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"code/util"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"

	"github.com/libp2p/go-libp2p"

	"github.com/adithyabhatkajake/libchatter/net"
	config "github.com/adithyabhatkajake/libsynchs/config"

	"code/msg"

	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
)

const (
	// ProtocolID is the ID for E2C Protocol
	ProtocolID = "synchs/proto/0.0.1"
	// ProtocolMsgBuffer defines how many protocol messages can be buffered
	ProtocolMsgBuffer = 100
)

// Init implements the Protocol interface
// 初始化实现协议接口
func (shs *SyncHS) Init(c *config.NodeConfig) {
	shs.NodeConfig = c
	shs.leader = DefaultLeaderID
	shs.view = 1 // View Number starts from 1 (convert view to round)
	shs.pendingCommands = make([][]byte, 1000)
	shs.timer = util.Timer{}
	// Setup maps
	shs.streamMap = make(map[uint64]*bufio.ReadWriter) //!!
	shs.cliMap = make(map[*bufio.ReadWriter]bool)      //!!
	shs.reputationMap = make(map[uint64]map[uint64]*big.Float)
	shs.voteMap = make(map[uint64]map[uint64]uint64)
	shs.proposalMap = make(map[uint64]map[uint64]uint64)
	shs.maliproposalMap = make(map[uint64]map[uint64]uint64)
	shs.equiproposalMap = make(map[uint64]map[uint64]uint64)
	shs.withproposalMap = make(map[uint64]map[uint64]uint64)
	shs.voteMaliMap = make(map[uint64]map[uint64]uint64)
	// shs.certBlockMap = make(map[*msg.BlockCertificate]chain.ExtBlock)
	shs.proposalByviewMap = make(map[uint64]*msg.Proposal)
	// Setup channels
	shs.msgChannel = make(chan *msg.SyncHSMsg, ProtocolMsgBuffer)
	shs.cmdChannel = make(chan []byte, ProtocolMsgBuffer)
	shs.voteChannel = make(chan *msg.Vote, ProtocolMsgBuffer)
	shs.proposeChannel = make(chan *msg.Proposal, ProtocolMsgBuffer)
	shs.eqEvidenceChannel = make(chan *msg.EquivocationEvidence, ProtocolMsgBuffer)
	shs.maliProEvidenceChannel = make(chan *msg.MalicousProposalEvidence, ProtocolMsgBuffer)
	shs.maliVoteEvidenceChannel = make(chan *msg.MalicousVoteEvidence, ProtocolMsgBuffer)
	shs.maliPropseChannel = make(chan *msg.Proposal, ProtocolMsgBuffer)
	shs.crossChannel = make(chan *msg.CrossTx, ProtocolMsgBuffer)
	// shs.blockCandidateChannel = make(chan *chain.Candidateblock, ProtocolMsgBuffer)
	shs.SyncChannel = make(chan bool, 1)
	shs.certMap = make(map[uint64]*msg.BlockCertificate)
	// Setup certificate for the first block
	shs.certMap[0] = &msg.GenesisCert
	shs.callFuncNotFinish = true
	shs.gcallFuncFinish = true
	shs.maliciousVoteInject = false
	shs.equivocatingProposalInject = false
	shs.withholdingProposalInject = false
	shs.maliciousProposalInject = false
	shs.maliVoteExists = false
	shs.maliProspoalExists = false
	shs.initialReplicaSore = new(big.Float).SetFloat64(1e-6)

}

// Setup sets up the network components 程序设置网络组件
func (shs *SyncHS) Setup(n *net.Network) error {
	shs.host = n.H
	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrStrings(shs.GetClientListenAddr()),
		libp2p.Identity(shs.GetMyKey()),
	)
	if err != nil {
		panic(err)
	}
	shs.pMap = n.PeerMap
	shs.cliHost = host
	shs.ctx = n.Ctx
	// Obtain a new chain 得到一条新链
	shs.bc = chain.NewChain()
	// TODO: create a new chain only if no chain is present in the data directory
	//仅当数据目录中不存在链时才创建新链
	// How to react to Protocol Messages 如何对协议消息做出反应
	shs.host.SetStreamHandler(ProtocolID, shs.ProtoMsgHandler)
	// How to react to Client Messages 如何对客户端消息做出反应
	shs.cliHost.SetStreamHandler(ClientProtocolID, shs.ClientMsgHandler)
	shs.shardDecision()

	//fmt.Println(shs.GetNumNodes())
	// Connect to all the other nodes talking E2C protocol 连接到所有其他使用E2C协议的节点
	wg := &sync.WaitGroup{} // For faster setup 为了更快的设置
	for idx, p := range shs.pMap {
		wg.Add(1)
		go func(idx uint64, p *peerstore.AddrInfo) {
			log.Trace("Attempting to open a stream with", p, "using protocol", ProtocolID)
			retries := 300
			for i := retries; i > 0; i-- {
				s, err := shs.host.NewStream(shs.ctx, p.ID, ProtocolID)
				if err != nil {
					log.Error("Error connecting to peers:", err)
					log.Info("Retry attempt ", retries-i+1, " to connect to node ", idx, " in a second")
					<-time.After(10 * time.Millisecond)
					continue
				}
				shs.netMutex.Lock()
				shs.streamMap[idx] = bufio.NewReadWriter( //流的存储 此时已经打开了此节点和其他所有节点之间的通讯流
					bufio.NewReader(s), bufio.NewWriter(s))
				shs.netMutex.Unlock()
				log.Info("Connected to Node ", idx)
				break
			}
			wg.Done()
		}(idx, p)
	}
	wg.Wait()
	log.Info("Setup Finished. Ready to do SMR:)", "The begin time is", time.Now())
	return nil
}

// start之前节点就已经打开对于每个节点的chanel了。除了客户端的，按道理来说应该也是打开了和客户端的通道。
// Start implements the Protocol Interface 开始协议，提议前的处理，投票的处理以及恶意投票的处理
func (shs *SyncHS) Start() {
	//First, start propsoal in forward step handler
	//首先，在正向步骤处理程序中启动 propsoal
	go shs.forwardProposalHandler()
	// Second, start vote handler concurrently 第二部启动投票处理程序
	go shs.voteHandler()
	go shs.crossHandler()
	//(start*)
	// Third, Start misbehavioushandler   开始行为等注入
	//go shs.MaliciousPropsoalHandler()
	//go shs.EquivocationEvidenceHandler()
	//go shs.MaliciousVoteEvidenceHandler()
	//go shs.MaliciousProposalEvidenceHandler()

	// Start E2C Protocol - Start message handler 启动信息处理handler
	shs.protocol() //接受到信息之后，去判断是提议还是其他信息
}

// ProtoMsgHandler reacts to all protocol messages in the network !!对网络中的所有协议消息做出反应！
func (shs *SyncHS) ProtoMsgHandler(s network.Stream) {
	// A global buffer to collect messages 用于收集消息的全局缓冲区
	buf := make([]byte, msg.MaxMsgSize)
	// Event Handler
	reader := bufio.NewReader(s)
	for {
		log.Trace("receive message")

		// Receive a message from anyone and process them 接收来自任何人的消息并处理它们
		len, err := reader.Read(buf) //将数据读入buf中
		if err != nil {
			panic(err)
			// return
		}
		// Use a copy of the message and send it to off for processing 使用邮件的副本并将其发送到 off 进行处理
		msgBuf := make([]byte, len)
		copy(msgBuf, buf[0:len])
		// React to the message in parallel and continue 并行响应消息并继续
		go shs.react(msgBuf)
	}
}
func (shs *SyncHS) shardDecision() {
	hostadd := shs.host.Addrs()
	host1 := hostadd[0]
	host11 := host1.String()
	host111 := strings.Split(host11, "/")
	hostto := host111[4] //这个hostto就是该节点的端口，我们根据端口号去分片
	//fmt.Println(hostto)
	//fmt.Println(len(shs.pMap))
	number, _ := strconv.Atoi(hostto)
	number = number % 10000
	if number < ((len(shs.pMap) + 1) / 2) {
		shs.shard = 1
		shs.leader = 0
	} else {
		shs.shard = 2
		shs.leader = uint64(((len(shs.pMap) + 1) / 2))
	}
	log.Info("成员ID：", shs.GetID(), "分片：", shs.shard, "领导：", shs.leader)
}
