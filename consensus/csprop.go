package consensus

import (
	"github.com/adithyabhatkajake/libsynchs/msg"
)

// 询问信息发送给分片1的leader 并且打开一个计时器，等待2个delta时间，若是还没有任何返回，则说明该领导是恶意的，我们需要执行共识协议去生产csvc-CC证书。
func (n *SyncHS) cspropfoward(msg1 *msg.SyncHSMsg) {
	quest := &msg.Quest{CrossTx: msg1.GetTx()}
	relayMsg := &msg.SyncHSMsg{}
	csprop1 := &msg.SyncHSMsg_Quest{Quest: quest}
	relayMsg.Msg = csprop1
	n.Broadcast4(relayMsg)

}
