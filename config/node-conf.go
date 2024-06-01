package config

import (
	"github.com/adithyabhatkajake/libchatter/crypto"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
	pb "google.golang.org/protobuf/proto"
)

// NodeConfig is an aggregation of all configs in one place
//NodeConfig 是所有配置在一个地方的聚合
type NodeConfig struct {
	// All Configs
	*SyncHSConfig //同步信息
	*NetConfig    //网络配置
	*CryptoConfig //加密配置
	//客户端口
	ClientPort string
	// PKI算法
	alg crypto.PKIAlgo
	// 我的私钥
	pvtKey crypto.PrivKey
	// 节点和公钥的映射
	nodeKeyMap map[uint64]crypto.PubKey
	// 缓存
	proto *NodeDataConfig
}

func (nc *NodeConfig) ToProto() *NodeDataConfig {
	if nc.proto == nil {
		nc.proto = &NodeDataConfig{
			ProtConfig: nc.SyncHSConfig,
			NetConfig:  nc.NetConfig,
			CryptoCon:  nc.CryptoConfig,
			ClientPort: nc.ClientPort,
		}
	}
	return nc.proto
}

func (nc *NodeConfig) FromProto(data *NodeDataConfig) {
	nc.SyncHSConfig = data.ProtConfig
	nc.NetConfig = data.NetConfig
	nc.CryptoConfig = data.CryptoCon
	nc.ClientPort = data.ClientPort
	nc.proto = data
}

// MarshalBinary implements Serializable interface
func (nc NodeConfig) MarshalBinary() ([]byte, error) {
	return pb.Marshal(nc.ToProto())
}

// UnmarshalBinary implements Deserializable interface
func (nc *NodeConfig) UnmarshalBinary(inpBytes []byte) error {
	data := &NodeDataConfig{}
	err := pb.Unmarshal(inpBytes, data)
	if err != nil {
		return nil
	}
	nc.FromProto(data)
	nc.init()
	return err
}

// MarshalJSON implements JSON Marshaller interface
// https://talks.golang.org/2015/json.slide#19
func (nc NodeConfig) MarshalJSON() ([]byte, error) {
	return jsonpb.Marshal(nc.ToProto())
}

// UnmarshalJSON implements Unmarshaller JSON interface
// https://talks.golang.org/2015/json.slide#19
func (nc *NodeConfig) UnmarshalJSON(bytes []byte) error {
	err := jsonpb.Unmarshal(bytes, nc.ToProto())
	if err != nil {
		return err
	}
	nc.init()
	return err
}

// init function initializes the structure assuming the config has been set
func (nc *NodeConfig) init() {
	alg, exists := crypto.GetAlgo(nc.GetKeyType())
	if exists == false {
		panic("Unknown key type")
	}
	nc.alg = alg
	nc.pvtKey = alg.PrivKeyFromBytes(nc.GetPvtKey())
	nc.nodeKeyMap = make(map[uint64]crypto.PubKey)
	for idx, pubkeyBytes := range nc.GetNodeKeyMap() {
		nc.nodeKeyMap[idx] = alg.PubKeyFromBytes(pubkeyBytes)
	}
}

// NewNodeConfig creates a NodeConfig object from NodeDataConfig
func NewNodeConfig(con *NodeDataConfig) *NodeConfig {
	nc := &NodeConfig{}
	nc.FromProto(con)
	nc.init()
	return nc
}
