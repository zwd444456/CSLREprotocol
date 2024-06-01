package consensus

import (
	"bufio"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/msg"
	"github.com/libp2p/go-libp2p-core/network"
	pb "google.golang.org/protobuf/proto"
)

// Implement how to talk to clients
const (
	ClientProtocolID = "synchs/client/0.0.1"
	Delta            = 8*time.Second + 0*time.Millisecond
)

// 将新客户端添加到 cliMap
func (n *SyncHS) addClient(rw *bufio.ReadWriter) {
	n.cliMutex.Lock() //客户端的锁
	n.cliMap[rw] = true
	n.cliMutex.Unlock()
}

// 移除一个client从climap中
func (n *SyncHS) removeClient(rw *bufio.ReadWriter) {
	// Remove rw from cliMap after disconnection
	n.cliMutex.Lock()
	delete(n.cliMap, rw)
	n.cliMutex.Unlock()
}

// 定义如何与客户端消息通信 n代表一个节点的所有信息
func (n *SyncHS) ClientMsgHandler(s network.Stream) {
	// A buffer to collect messages 用于收集消息的缓冲区
	buf := make([]byte, msg.MaxMsgSize)
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	// Add client for later contact 添加客户端以供以后联系
	n.addClient(rw)
	//inital reputationMap for all nodes 所有节点的初始信誉映射
	//n.initialReputationMap() //我的协议不需要信誉度相关的东西
	//上面的是初始化节点之间互相连接之后，将客户和节点之间连接并添加到节点的客户中，然后初始化节点对于每个节点的信誉度
	// Set timer for all nodes 为所有节点设置计时器 不懂为什么会切换试图
	//n.setConsensusTimer()
	log.Debug("finish the setting of timer")
	// Event Handler 事件处理程序 处理接受到的信息并将其放到悬挂命令之后。
	for {
		// Receive a message from a client and process them 从客户端接收消息并处理它们
		len, err := rw.Read(buf)
		if err != nil {
			log.Error("Error receiving a message from the client-", err)
			return
		}
		// Send a copy for reacting 发送副本以供响应
		inMsg := &msg.SyncHSMsg{}
		err = pb.Unmarshal(buf[0:len], inMsg) //将收到的信息解码放到syncHs中
		if err != nil {
			log.Error("Error unmarshalling cmd from client")
			log.Error(err)
			continue
		}
		var cmd []byte
		if cmd = inMsg.GetTx(); cmd == nil {
			log.Error("invalid command received from client")
			continue
		}
		fmt.Println(cmd)
		fmt.Println(string(inMsg.GetTx()[:]))
		if strings.Contains(string(inMsg.GetTx()[:]), "c") {
			fmt.Println("it is cross_tx")
			log.Info(time.Now())
			go n.crossCommands(cmd)
			continue
		}
		// Add command
		// log.Debug("now round is", n.view)添加 Cmds 和启动计时器（如果有足够的命令）
		go n.addCmdsAndStartTimerIfSufficientCommands(cmd)
	}
}

// ClientBroadcast  向此实例已知的所有客户端发送协议消息
func (n *SyncHS) ClientBroadcast(m *msg.SyncHSMsg) {
	data, err := pb.Marshal(m) //设置为传输的格式
	if err != nil {
		log.Error("Failed to send message", m, "to client")
		panic(err)
	}
	n.cliMutex.Lock()
	defer n.cliMutex.Unlock()
	for cliBuf := range n.cliMap {
		log.Trace("Sending to", cliBuf)
		cliBuf.Write(data)
		cliBuf.Flush()
	}
	log.Trace("Finish client broadcast for", m)
}
func (n *SyncHS) setConsensusTimer() {
	n.timer.SetCallAndCancel(n.callback)
	//+ 150*time.Millisecond
	n.timer.SetTime(Delta)
}

// 反馈信息 提交成功了。反馈这个区块是否成功还是失败。
func (n *SyncHS) callback() {

	log.Debug(n.GetID(), "callbackFuncation have been prepared!", time.Now())
	if n.withholdingProposalInject { //如果有withhoulding攻击注入，则 转换试图
		log.Info("In round", n.view, "withholding block have been detected")
		//Handle withholding behaviour
		//计算前一轮的信誉度
		n.addNewViewReputaiontoMap()
		synchsmsg := &msg.SyncHSMsg{}
		ack := &msg.SyncHSMsg_Ack{} //提交信息
		log.Info("Committing an emptyblock in withholding case-", n.view)
		log.Info("The block commit time is", time.Now())
		log.Debug(n.GetID(), "node Blockchain height and view number is", n.bc.Head, "AND", n.view)
		// Let the client know that we committed this block 让客户端知道我们提交了此块
		ack.Ack = &msg.CommitAck{
			Block: n.bc.BlocksByHeight[n.bc.Head].ToProto(),
		}
		synchsmsg.Msg = ack
		// Tell all the clients, that I have committed this block 告诉所有客户端，我已经提交了这个块
		n.ClientBroadcast(synchsmsg)
		n.view++
		n.changeLeader()
		n.SyncChannel <- true
		log.Debug(len(n.SyncChannel), "the next leader is", n.leader)
		return
	}
	if n.equivocatingProposalInject { //如果有模棱两可提议
		//this equivocation behaviour have been handle
		n.equivocatingProposalInject = false
		log.Info("In round", n.view, "equivocating block have been detected")
		n.addNewViewReputaiontoMap()
		synchsmsg := &msg.SyncHSMsg{}
		ack := &msg.SyncHSMsg_Ack{}
		log.Info("Committing an emptyblock in equivocation case-", n.view)
		log.Info("The block commit time is", time.Now())
		log.Debug(n.GetID(), "node Blockchain height and view number is", n.bc.Head, "AND", n.view)
		// Let the client know that we committed this block
		ack.Ack = &msg.CommitAck{
			Block: n.bc.BlocksByHeight[n.bc.Head].ToProto(),
		}
		synchsmsg.Msg = ack
		// Tell all the clients, that I have committed this block
		n.ClientBroadcast(synchsmsg) //将此消息也广播到所有的客户端 将提交信息给客户端
		n.view++
		n.changeLeader()
		n.SyncChannel <- true
		log.Debug(len(n.SyncChannel))
		return
	}
	n.addNewViewReputaiontoMap()
	synchsmsg := &msg.SyncHSMsg{}
	ack := &msg.SyncHSMsg_Ack{}
	//
	log.Debug(n.GetID(), "node Blockchain height and view number is", n.bc.Head, "AND", n.view)
	//根据试图信息找到裁断证书
	_, exist := n.getCertForBlockIndex(n.view)
	if !exist {
		log.Debug("fail to generate certificate in round", n.view)
		n.SyncChannel <- true
		return
	}

	log.Info("Committing an correct block-", n.view)
	log.Info("The block commit time of ", n.GetID(), "is", time.Now())
	ack.Ack = &msg.CommitAck{
		Block: n.bc.BlocksByHeight[n.bc.Head].ToProto(),
	}
	synchsmsg.Msg = ack

	// Tell all the clients, that I have committed this block
	n.ClientBroadcast(synchsmsg)
	n.view++
	n.changeLeader()
	n.SyncChannel <- true
	log.Debug(len(n.SyncChannel))

}

// 初始化信誉值 n.reputationMap[n.view]=0:1e-06 1:1e-06 2:1e-06
func (n *SyncHS) initialReputationMap() {
	n.repMapLock.Lock()
	defer n.repMapLock.Unlock()
	log.Debug(n.pMap)
	for i := uint64(0); i <= uint64(len(n.pMap)); i++ {
		// n.reputationMapwithoutRound[i] = n.initialReplicaSore
		if _, exists := n.reputationMap[n.view]; exists {
			n.reputationMap[n.view][i] = n.initialReplicaSore

		} else {
			n.reputationMap[n.view] = make(map[uint64]*big.Float)
			n.reputationMap[n.view][i] = n.initialReplicaSore
		}
	}

	log.Debug("finish repmap setting with Node", n.GetID(), "'s repmap is", n.reputationMap[n.view])
	//给他们的信誉值初始为0 相当于 信誉值：
}
