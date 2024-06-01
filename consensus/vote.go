package consensus

import (
	"math/big"

	"github.com/adithyabhatkajake/libchatter/crypto"
	"github.com/adithyabhatkajake/libchatter/log"
	"github.com/adithyabhatkajake/libsynchs/chain"
	msg "github.com/adithyabhatkajake/libsynchs/msg"
	pb "google.golang.org/protobuf/proto"
)

// Vote channel //TODO malicious vote and equivocation proposal detected!
// TODO{convert the setting of certificate to reputation setting}
// 投票通道TODO恶意投票和模棱两可的提案被检测到！TODO{将证书设置转换为信誉设置}
func (n *SyncHS) voteHandler() {
	// Map leader to a map from sender to its vote
	voteMap := make(map[uint64]map[uint64]*msg.Vote)
	// pendingVotes := make(map[crypto.Hash][]*msg.Vote)
	isCertified := make(map[uint64]bool)
	myID := n.GetID()
	// Faults := n.GetNumberOfFaultyNodes()
	for {
		v, ok := <-n.voteChannel
		if !ok {
			log.Error("Vote channel error")
			continue
		}
		// if len(n.voteChannel) <= 0 {
		// 	log.Trace("NODES DON'T ACCESS TO VOTESTEP WAIT")
		// 	continue
		// }
		if v.Owner != n.leader {
			log.Warn(v.GetVoter(), "'s Malicious vote have been detected.", "Owner is", v.Owner, "leader is", n.leader)
			go n.sendMalivoteEvidence(v)
			continue
		}
		bhash := crypto.ToHash(v.GetBlockHash())
		blk := n.getBlock(bhash)
		height := n.view
		view := v.GetView()
		_, exists := voteMap[height]
		if !exists {
			voteMap[height] = make(map[uint64]*msg.Vote)
			isCertified[height] = false
		}

		_, exists = voteMap[height][v.GetVoter()]
		if exists {
			log.Debug("Duplicate vote received")
			continue
		}

		// My vote is always valid
		if v.GetVoter() != myID {
			isValid := n.isVoteValid(v, blk)
			if !isValid {
				log.Error("Invalid vote message")
				continue
			}
		}
		n.addVotetoMap(v.ToProto())
		//only change votemap for reputation
		isCert, exists := isCertified[height]
		if exists && isCert {
			// This vote for this block is already certified, ignore 此块的投票已经通过认证，请忽略
			continue
		}
		voteMap[height][v.GetVoter()] = v
		// if uint64(len(voteMap[height])) <= Faults {
		// 	log.Debug("Not enough votes to build a certificate")
		// 	continue
		// }

		//reutation version!!TODO
		var repSumInCert *big.Float = new(big.Float).SetFloat64(0)
		for i := range voteMap[height] {
			repSumInCert = repSumInCert.Add(repSumInCert, n.reputationMap[n.view][i])
		}
		// log.Debug("OUTPUT SCORE", repSumInCert)
		if repSumInCert.Cmp(n.GetCertBenchMark(n.view)) == -1 || repSumInCert.Cmp(n.GetCertBenchMark(n.view)) == 0 {
			log.Debug("Not enough reputation to build a certificate")
			continue
		}
		log.Debug("Building a certificate")

		// Our certificate for height v.Data.Block.Data.Index is ready now 我们的高度证书v.Data.Block.Data.Index现已准备就绪
		bcert := NewCert(voteMap[height], bhash, view)
		isCertified[height] = true
		// need an anthoer map for non-leader node get exblock  需要非领导者节点的花旗映射获取exblock
		go func() {
			//add this vote to votemap
			n.addCert(bcert, height)
			// if n.leader == myID {
			// 	n.propose()
			// }
		}()
	}
}

func (n *SyncHS) isVoteValid(v *msg.Vote, blk *chain.ExtBlock) bool {
	data, err := pb.Marshal(v.ProtoVoteData)
	if err != nil {
		log.Error("Error marshalling vote data")
		log.Error(err)
		return false
	}
	sigOk, err := n.GetPubKeyFromID(v.GetVoter()).Verify(data, v.GetSignature())
	if err != nil {
		log.Error("Vote Signature check error")
		log.Error(err)
		return false
	}
	if !sigOk {
		log.Error("Invalid vote Signature")
		return sigOk
	}
	return sigOk
}

// change it and we can found the miner of the block in vote 更改它，我们可以在投票中找到区块的矿工 确定有足够的提议，节点为该区块投票
func (n *SyncHS) voteForBlock(exprop *msg.ExtProposal) {
	log.Info("NODE", n.GetID(), "is voting for", exprop.Miner, "'s block")
	pvd := &msg.ProtoVoteData{
		BlockHash: exprop.ExtBlock.GetBlockHash().GetBytes(),
		View:      n.view,
		Owner:     exprop.Miner,
	}
	data, err := pb.Marshal(pvd)
	if err != nil {
		log.Error("Error marshing vote data during voting")
		log.Error(err)
		return
	}
	sig, err := n.GetMyKey().Sign(data)
	if err != nil {
		log.Error("Error signing vote")
		log.Error(err)
		return
	}
	pvb := &msg.ProtoVoteBody{
		Voter:     n.GetID(),
		Signature: sig,
	}
	pv := &msg.ProtoVote{
		Body: pvb,
		Data: pvd,
	}
	voteMsg := &msg.SyncHSMsg{}
	voteMsg.Msg = &msg.SyncHSMsg_Vote{Vote: pv}
	v := &msg.Vote{}
	v.FromProto(pv)
	n.addVotetoMap(pv)
	go func() { //投票完了并广播
		//the voter change his voteMap by himself
		// Send vote to all the nodes
		n.Broadcast(voteMsg)
		// Handle my own vote
		n.voteChannel <- v
	}()
}
func (n *SyncHS) addMaliVotetoMap(v *msg.ProtoVote) {
	// n.voteMaliLock.Lock()
	if _, exists := n.voteMaliMap[n.view]; exists {
		n.voteMaliMap[n.view][v.GetBody().Voter] = 1
	} else {
		n.voteMaliMap[n.view] = make(map[uint64]uint64)
		n.voteMaliMap[n.view][v.GetBody().Voter] = 1
	}
	// n.voteMaliLock.Unlock()
}

// 添加投票映射
func (n *SyncHS) addVotetoMap(v *msg.ProtoVote) {
	// n.voteMapLock.Lock()
	// n.voterMap[v.GetBody().Voter] = 1
	if _, exists := n.voteMap[n.view]; exists {
		n.voteMap[n.view][v.GetBody().Voter] = 1
	} else {
		n.voteMap[n.view] = make(map[uint64]uint64)
		n.voteMap[n.view][v.GetBody().Voter] = 1
	}
	// log.Debug("Node", n.GetID(), "'S votemap is", n.voteMap)
	// log.Debug("vote", n.voteMap)
	// n.voteMapLock.Unlock()
}

func (n *SyncHS) GetCertBenchMark(viewNum uint64) *big.Float {
	var totalReputationScore *big.Float = new(big.Float).SetFloat64(0)
	var certBenchMark *big.Float = new(big.Float).SetFloat64(0.5)
	for _, v := range n.reputationMap[viewNum] {
		totalReputationScore = totalReputationScore.Add(totalReputationScore, v)
	}
	certBenchMarkScore := new(big.Float).Mul(totalReputationScore, certBenchMark)
	return certBenchMarkScore

}
