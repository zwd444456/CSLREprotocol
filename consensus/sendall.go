package consensus

import (
	"code/msg"
	"github.com/adithyabhatkajake/libchatter/log"
	pb "google.golang.org/protobuf/proto"
	"time"
)

// Broadcast broadcasts a protocol message to all the nodes!! 向所有节点广播协议消息！！
func (n *SyncHS) Broadcast(m *msg.SyncHSMsg) error {

	n.netMutex.Lock()

	defer n.netMutex.Unlock()
	data, err := pb.Marshal(m)
	if err != nil {
		return err
	}
	// If we fail to send a message to someone, continue
	for idx, s := range n.streamMap {
		_, err = s.Write(data)
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
			continue
		}
		err = s.Flush()
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
		}
	}
	return nil
}

// 广播给输出分片2的的所有成员
func (n *SyncHS) Broadcast2(m *msg.SyncHSMsg) error {
	n.netMutex.Lock()
	defer n.netMutex.Unlock()
	data, err := pb.Marshal(m)
	if err != nil {
		return err
	}
	// 只传送给输出分配的leader
	log.Info("我是输出分片的leader，我在规定时间内没收到来自shard1的裁断证书，所以我要通知我的分片成员，让他们去询问shard1的领导是什么情况")
	for idx, s := range n.streamMap {
		if idx < n.GetID() {
			continue
		}
		_, err = s.Write(data)
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
			continue
		}
		err = s.Flush()
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
		}
	}
	return nil
}

// 发送给输入分片和输出分片的leader
func (n *SyncHS) Broadcast4(m *msg.SyncHSMsg) error {
	n.netMutex.Lock()
	defer n.netMutex.Unlock()
	data, err := pb.Marshal(m)
	if err != nil {
		return err
	}
	// 只传送给输出分配的leader
	log.Info("我是输出分片的成员，我的leader告诉我没有收到输入分片的裁断证书，所以我再次去问询输入分片shard1的leader索要该证书信息")
	for idx, s := range n.streamMap {
		if idx != 0 && idx != n.GetNumNodes()/2 {
			continue
		}
		_, err = s.Write(data)
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
			continue
		}
		err = s.Flush()
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
		}
	}
	return nil
}

// 广播给第二个分片的领导
func (n *SyncHS) Broadcast1(m *msg.SyncHSMsg) error {
	n.netMutex.Lock()
	defer n.netMutex.Unlock()
	data, err := pb.Marshal(m)
	if err != nil {
		return err
	}
	// 只传送给输出分配的leader

	for idx, s := range n.streamMap {
		if idx != n.GetNumNodes()/2 {
			continue
		}
		_, err = s.Write(data)
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
			continue
		}
		err = s.Flush()
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
		}
	}
	return nil
}

// 广播切换领导的信息 给输入分片1的所有成员
func (n *SyncHS) Broadcast3(m *msg.SyncHSMsg) error {
	n.netMutex.Lock()
	defer n.netMutex.Unlock()
	data, err := pb.Marshal(m)
	if err != nil {
		return err
	}
	// 只传送给输出分配的leader
	log.Info("我是输出分片，我们进行了共识，当前要让领导去发送转换证书给shard1的所有成员")
	log.Info(time.Now())
	for idx, s := range n.streamMap {
		if idx != 1 {
			continue
		}
		_, err = s.Write(data)
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
			continue
		}
		err = s.Flush()
		if err != nil {
			log.Error("Error while sending to node", idx)
			log.Error("Error:", err)
		}
	}
	return nil
}

// Equivocating block (broadcast)
func (n *SyncHS) EquivocatingBroadcast(m1 *msg.SyncHSMsg, m2 *msg.SyncHSMsg) (e1, e2 error) {
	n.netMutex.Lock()
	defer n.netMutex.Unlock()
	data1, err1 := pb.Marshal(m1)
	data2, err2 := pb.Marshal(m2)
	if err1 != nil || err2 != nil {
		return err1, err2
	}
	// If we fail to send a message to someone, continue
	for idx, s := range n.streamMap {
		//example for 3node is id == 0
		if idx%2 == 0 {
			_, err1 = s.Write(data1)
			if err1 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err1)
				continue
			}
			err1 = s.Flush()
			if err1 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err1)
			}
		} else {
			_, err2 = s.Write(data2)
			if err2 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err2)
				continue
			}
			err2 = s.Flush()
			if err2 != nil {
				log.Error("Error while sending to node", idx)
				log.Error("Error:", err2)
			}

		}

	}
	return nil, nil

}
