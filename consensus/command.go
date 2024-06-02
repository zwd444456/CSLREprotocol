package consensus

import (
	"code/msg"
	"github.com/adithyabhatkajake/libchatter/log"
	time2 "time"
)

func (n *SyncHS) crossCommands(cmd []byte) {
	if n.shard == 1 && n.leader == n.GetID() {
		//TODO:发送一条消息给输出分片的leader
		log.Info("我是第一个输入分片的领导，我将向输出分片的领导ID为3的人发送跨交易的裁断证书。但是我是恶意的，所以此处我们在生成可用性证书之后，我并不准备发送该信息")
		//n.sendAccessCert(cmd)
	}
	if n.shard == 2 && n.leader == n.GetID() {
		log.Info("我是输出分片的领导，我现在会等待一个delta时间直到收到所有的输入的可用性证书。")
		time := time2.NewTimer(Delta)
		select {
		case <-time.C:
			log.Info("一个delta时间到了，没有收到裁断证书，所以我需要通知我的成员们。让他们去询问该分片的领导是否忘记发送了证书")
			log.Info(time2.Now())
			n.csvc_propose(cmd)
		case msg := <-n.crossChannel:
			log.Info("我们收到了裁断证书，所以不会进行协议去更换领导")
			log.Info(msg)
			if !time.Stop() {
				<-time.C
			}
		}
		//等待一个代尔塔，没等到则
	}
}
func (n *SyncHS) csvc_propose(cmd []byte) {
	csProp := &msg.CsvcPropose{
		CrossTx: cmd,
		Shard:   1,
	}
	relayMsg := &msg.SyncHSMsg{}
	csprop1 := &msg.SyncHSMsg_CsProp{CsProp: csProp}
	relayMsg.Msg = csprop1
	n.Broadcast2(relayMsg)
}

// 添加 Cmds 和启动计时器（如果有足够的命令）
func (n *SyncHS) addCmdsAndStartTimerIfSufficientCommands(cmd []byte) {
	// log.Debug("procedure in this step in round", n.view)
	n.cmdMutex.Lock()
	n.pendingCommands = append(n.pendingCommands, cmd) //将命令添加到悬挂起来的命令最后
	n.cmdMutex.Unlock()
	// n.pendingCommands.PushBack(cmd)
	//&& n.GetID() == n.leader
	if uint64(len(n.pendingCommands)) >= n.GetBlockSize() { //悬挂命令数量大于了当前块所能存放的交易数。
		if n.gcallFuncFinish { //是否要生产攻击注入
			//n.startConsensusTimerWithEquivocation()
			// n.startConsensusTimerWithWithhold()
			//n.startConsensusTimerWithMaliciousPropsoal()
			n.startConsensusTimer()
			//n.startConsensusTimerWithMaliciousVote() //启动恶意投票共识
			n.gcallFuncFinish = false //启动完成，不需要在启动了。
		}
		//16
		if len(n.SyncChannel) == 1 {
			<-n.SyncChannel
			// log.Info("node", n.GetID(), "'s pendingCommands len is", len(n.pendingCommands))
			// time.Sleep(time.Second * 2)
			//n.startConsensusTimerWithEquivocation()
			//n.startConsensusTimerWithMaliciousPropsoal()
			//n.startConsensusTimerWithMaliciousVote() //如果n.syncchannel通道的内容只有一个。则启动恶意投票共识
			n.startConsensusTimer()
			// n.startConsensusTimerWithWithhold()
		}
	}
}

// !!TODO all pengdingCommands should be change not only  leader?所有pengdingCommands都应该改变而不仅仅是领导者吗？
func (n *SyncHS) getCmdsIfSufficient() ([][]byte, bool) {
	blkSize := n.GetBlockSize()
	n.cmdMutex.Lock()
	defer n.cmdMutex.Unlock()
	numCmds := uint64(len(n.pendingCommands))
	if numCmds < blkSize {
		return nil, false
	}
	cmds := make([][]byte, blkSize)
	// Copy slice blkSize commands from pending Commands
	copy(cmds, n.pendingCommands[numCmds-blkSize:])
	// Update old slice
	n.pendingCommands = n.pendingCommands[:numCmds-blkSize]
	return cmds, true
}
