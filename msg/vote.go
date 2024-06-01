package msg

// ToProto returns a protocol buffer message for external communication of Vote 返回用于 Vote 的外部通信的协议缓冲区消息
func (v *Vote) ToProto() *ProtoVote {
	val := v.proto.Load()
	if val == nil {
		pv := &ProtoVote{
			Data: v.ProtoVoteData,
			Body: v.ProtoVoteBody,
		}
		v.proto.Store(pv)
		return pv
	}
	return val.(*ProtoVote)
}

// FromProto updates a vote structure with the contents of a protocol buffer 使用协议缓冲区的内容更新投票结构
func (v *Vote) FromProto(data *ProtoVote) {
	v.VoteData = &VoteData{
		ProtoVoteData: data.GetData(),
	}
	v.VoteBody = &VoteBody{
		ProtoVoteBody: data.GetBody(),
	}
	v.proto.Store(data)
}
