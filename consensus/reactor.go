package consensus

import (
	"github.com/adithyabhatkajake/libchatter/log"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "google.golang.org/protobuf/proto"
	"time"
)

// 对收到的消息进行处理
func (n *SyncHS) react(m []byte) error {
	log.Trace("Received a message of size", len(m))
	inMessage := &msg.SyncHSMsg{}
	// n.netMutex.Lock()
	err := pb.Unmarshal(m, inMessage) //解码
	log.Info(time.Now())
	// n.netMutex.Unlock()
	if err != nil {
		log.Error("Received an invalid protocol message", err)
		return err
	} //将收到的信息通过msgchannel通道传到节点
	n.msgChannel <- inMessage
	log.Trace("there are no error")
	return nil
}

// 对于收到的信息去做提议和转发等
func (n *SyncHS) protocol() {
	// Process protocol messages 处理协议消息
	for {
		msgIn, ok := <-n.msgChannel
		if !ok {
			log.Error("Msg channel error")
			return
		}
		log.Trace("Received msg", msgIn.String())
		switch x := msgIn.Msg.(type) {
		case *msg.SyncHSMsg_Cross: //正常执行
			if uint64(n.GetID()) == n.leader {
				if n.shard == 2 {
					cross := msgIn.GetCross()
					log.Info("我是第二个分片委员会的leader，所以我接受了来自分片1领导的裁断证书")
					go func() {
						n.crossChannel <- cross
					}()
				}
			}
		case *msg.SyncHSMsg_CsProp:
			log.Info("我收到了领导对于某条交易信息的分片领导的问题，所以我需要去询问该领导")
			go n.cspropfoward(msgIn)
		case *msg.SyncHSMsg_Quest:

			go n.s2leadersend()
		case *msg.SyncHSMsg_Csvc_CC:
			log.Info("我是分片1的成员，我收到来自分片2的领导发送的切换leader的信息，并且是经过他的委员会诚实大多数的签名去接受的，所以我们内部现在会进行试图转换协议去切换领导")
			log.Info(time.Now())
		case *msg.SyncHSMsg_Prop:
			prop := msgIn.GetProp()
			if prop.ForwardSender == n.leader {
				if prop.GetMiner() == n.leader {
					log.Debug("Received a proposal from ", prop.GetMiner())
					// Send proposal to forward step
					go n.forward(prop)
				} else {
					log.Debug("Received a Malicious proposal from ", prop.GetMiner(), "in round ", n.view)
					go func() {
						n.maliPropseChannel <- prop
					}()
				}
			} else {
				n.proposeChannel <- prop
			}
			//(start*)
		// case *msg.SyncHSMsg_Eqevidence:
		// 	eqEvidence := msgIn.GetEqevidence()
		// 	go func() {
		// 		n.eqEvidenceChannel <- eqEvidence
		// 	}()

		// case *msg.SyncHSMsg_Mpevidence:
		// 	maliProEvidence := msgIn.GetMpevidence()
		// 	go func() {
		// 		n.maliProEvidenceChannel <- maliProEvidence
		// 	}()
		case *msg.SyncHSMsg_Mvevidence:
			maliVoteEvidence := msgIn.GetMvevidence()
			go func() {
				n.maliVoteEvidenceChannel <- maliVoteEvidence
			}()
		case *msg.SyncHSMsg_Vote:
			pvote := msgIn.GetVote()
			vote := &msg.Vote{}
			vote.FromProto(pvote)
			go func() {
				n.voteChannel <- vote
			}()

		case nil:
			log.Warn("Unspecified msg type", x)
		default:
			log.Warn("Unknown msg type", x)
		}
	}
}
