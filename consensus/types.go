package consensus

import (
	"bufio"
	"context"
	"math/big"
	"sync"

	chain "github.com/adithyabhatkajake/libsynchs/chain"
	config "github.com/adithyabhatkajake/libsynchs/config"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	lutil "github.com/adithyabhatkajake/libsynchs/util"

	"github.com/libp2p/go-libp2p-core/host"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
)

// SyncHS implements the consensus protocol!!
// 同步HS实现了共识协议 该同步hs实现了区块链中的共识协议。
type SyncHS struct {
	// Network data structures 网络通讯的数据结构
	host    host.Host
	cliHost host.Host
	ctx     context.Context
	shard   uint64
	// Maps 映射
	// Mapping between ID and libp2p-peer id和libp2p-peer的映射
	pMap map[uint64]*peerstore.AddrInfo
	// A set of all known clients //一组所有已知客户端
	cliMap map[*bufio.ReadWriter]bool
	// A map of node ID to its corresponding RW stream 节点 ID 到其相应 读写 流的映射
	streamMap map[uint64]*bufio.ReadWriter
	// A map of hash to pending commands 哈希到挂起命令的映射
	pendingCommands [][]byte
	// A mapping between the block number to its commit timer 块到他的提交时间的映射
	// timerMaps map[uint64]*lutil.Timer //
	// Certificate map  证书映射
	certMap map[uint64]*msg.BlockCertificate
	// correct vote map (recorder:origin:the number/value of vote/proposal/reputation)
	// 正确投票到（记录者：来源：投票提案的编号值声誉）的映射
	voteMap map[uint64]map[uint64]uint64
	// malicous vote map 恶意投票的映射
	voteMaliMap map[uint64]map[uint64]uint64
	// correct proposal map
	//正确提议的映射
	proposalMap map[uint64]map[uint64]uint64
	//equivocate proposal map
	//模棱两可提议的映射
	equiproposalMap map[uint64]map[uint64]uint64
	//withholding proposal map
	// withholding 提议的映射
	withproposalMap map[uint64]map[uint64]uint64
	//malicious proposal map
	//恶意提议的映射
	maliproposalMap map[uint64]map[uint64]uint64
	//Reputation map
	//信誉的映射
	reputationMap map[uint64]map[uint64]*big.Float
	//ProosalByheightMap
	//对于试图提议的映射
	proposalByviewMap map[uint64]*msg.Proposal
	/* Locks - We separate all the locks, so that acquiring
	one lock does not make other goroutines stop */
	/*//锁 - 我们分离所有锁，以便获取一个锁不会使其他 goroutines 停止*/
	crossLock          sync.RWMutex //THe lock to modify 大大大
	cliMutex           sync.RWMutex // The lock to modify cliMap 锁以修改climap
	netMutex           sync.RWMutex // The lock to modify streamMap:使用网络流与其他节点通信时使用互斥锁
	cmdMutex           sync.RWMutex // The lock to modify pendingCommands 修改挂起指令
	timerLock          sync.RWMutex // The lock to modify timerMaps 时间映射锁
	certMapLock        sync.RWMutex // The lock to modify certMap  证书映射锁
	leaderByviewLock   sync.RWMutex //The lock to modify leaderroundMap  领导周次锁
	repMapLock         sync.RWMutex // The lock to modify reputationMap
	voteMapLock        sync.RWMutex // The lock to modify reputationMap
	propMapLock        sync.RWMutex // The lock to modify reputationMap
	malipropLock       sync.RWMutex //........
	equipropLock       sync.RWMutex //
	voteMaliLock       sync.RWMutex
	withpropoLock      sync.RWMutex
	proposalByviewLock sync.RWMutex
	certBlockLock      sync.RWMutex

	// Channels 通道
	questChannel            chan *msg.Quest
	crossChannel            chan *msg.CrossTx
	msgChannel              chan *msg.SyncHSMsg                // All messages come here first所有消息首先到达此处
	cmdChannel              chan []byte                        // All commands are re-directed here 所有命令都重定向到此处
	voteChannel             chan *msg.Vote                     // All votes are sent here 所有投票发送到这
	SyncChannel             chan bool                          //make a channel to store the signal of timerfinish制作一个通道来存储计时器的信号
	proposeChannel          chan *msg.Proposal                 // All proposals are sent here 所有的提议发送到这
	eqEvidenceChannel       chan *msg.EquivocationEvidence     // All EquivocationEvidence are sent here  所有模棱两可的行为发送到这
	maliProEvidenceChannel  chan *msg.MalicousProposalEvidence //....所有恶意提议行为
	maliVoteEvidenceChannel chan *msg.MalicousVoteEvidence     //....所有恶意投票行为
	maliPropseChannel       chan *msg.Proposal                 //....所有提议行为
	// errCh          chan error          // All errors are sent here

	// Block chain 区块链
	bc *chain.BlockChain

	// Protocol information 提议信息
	leader uint64
	view   uint64 //n的试图

	// Embed the config 嵌入配置
	*config.NodeConfig

	// The varible of attack injection 攻击注入的变种
	equivocatingProposalInject bool //模棱两可提议的注入
	withholdingProposalInject  bool //withholding提议注入
	maliciousProposalInject    bool //恶意提议注入
	maliciousVoteInject        bool //恶意投票提议
	//check malicious propsoal/vote recording 检查恶意投票和提议的记录
	maliProspoalExists bool //是否存在
	maliVoteExists     bool //是否存在
	// Check callfunc state 检查呼叫函数状态
	callFuncNotFinish bool
	gcallFuncFinish   bool
	// The timer of every node 每个节点的时间
	timer  lutil.Timer
	timer2 lutil.Timer //跨分片的节点时间
	//initial reputaion of all nodes 初始化所有节点的信誉
	initialReplicaSore *big.Float
}
