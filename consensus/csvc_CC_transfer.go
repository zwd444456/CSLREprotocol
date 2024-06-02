package consensus

import (
	"code/msg"
	"github.com/adithyabhatkajake/libchatter/log"
)

func (n *SyncHS) Crosscertforward() {
	crossCert := &msg.CrossShardViewChangeCommitCertificate{
		Shardleader: 0,
		Signatures:  nil,
	}
	relayMsg := &msg.SyncHSMsg{}
	csprop1 := &msg.SyncHSMsg_Csvc_CC{Csvc_CC: crossCert}
	relayMsg.Msg = csprop1
	log.Info("我们进行完了共识,并且发送切换证书信息")
	n.Broadcast3(relayMsg)

}
