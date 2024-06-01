package consensus

import (
	"fmt"
	"github.com/adithyabhatkajake/libchatter/log"
)

func (n *SyncHS) crossHandler() {
	for {
		v, ok := <-n.crossChannel
		if !ok {
			log.Error("cross channel error")
			continue
		}
		fmt.Println(v)
		log.Info("我们正常收到的所有的裁断证书，所以正常执行p2p的下一阶段。")
	}
}
