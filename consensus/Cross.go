package consensus

import (
	"code/msg"
	"github.com/adithyabhatkajake/libchatter/log"
	time2 "time"
)

func (n *SyncHS) sendAccessCert(cmd []byte) {
	cmd[0] = 'c'
	cross := &msg.CrossTx{
		CrossTx: cmd,
		Id:      n.GetID(),
	}
	relayMsg := &msg.SyncHSMsg{}
	cross_tx := &msg.SyncHSMsg_Cross{Cross: cross}
	relayMsg.Msg = cross_tx
	go func() {
		//Change itself proposal map
		// n.addProposaltoMap()
		// Leader sends new block to all the other nodes
		n.Broadcast1(relayMsg) //将新的区块发送到所有其他节点
	}()
}
func (n *SyncHS) s2leadersend() {
	if n.GetID() == 0 {
		log.Info("因为我们模拟的是恶意领导，所以此处我们是不会返回任何东西的。")
	}
	if n.GetID() == uint64(n.GetNumNodes()/2) {
		isit := n.waitTime()
		if isit == true {
			//执行内部共识协议，此处大概估摸一个时间
			time2.Sleep(2 * Delta)
			n.Crosscertforward()
		} else {
			log.Info("shoudaole suoyibufa ")
		}
	}
}
func (n *SyncHS) waitTime() bool {
	time := time2.NewTimer(2 * Delta)

	select {
	case <-time.C:
		log.Info("两个delta时间到了，没有收到裁断证书，所以我们输出分片2需要去提出跨试图转换协议然后去更换领导。")
		return true
	case msg := <-n.questChannel:
		log.Info("我们收到了裁断证书，所以不会进行内部共识去更换领导")
		log.Info(msg)
		return false
		if !time.Stop() {
			<-time.C
		}
	}
	return false
}
