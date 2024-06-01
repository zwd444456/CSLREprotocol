// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: config/protocol.proto

package config

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ProtoInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeSize  uint64 `protobuf:"varint,1,opt,name=NodeSize,proto3" json:"NodeSize,omitempty"`   // Number of nodes in the system
	Faults    uint64 `protobuf:"varint,2,opt,name=Faults,proto3" json:"Faults,omitempty"`       // Number of Faults in the system
	BlockSize uint64 `protobuf:"varint,3,opt,name=BlockSize,proto3" json:"BlockSize,omitempty"` // Number of commands in a block
}

func (x *ProtoInfo) Reset() {
	*x = ProtoInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_protocol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoInfo) ProtoMessage() {}

func (x *ProtoInfo) ProtoReflect() protoreflect.Message {
	mi := &file_config_protocol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoInfo.ProtoReflect.Descriptor instead.
func (*ProtoInfo) Descriptor() ([]byte, []int) {
	return file_config_protocol_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoInfo) GetNodeSize() uint64 {
	if x != nil {
		return x.NodeSize
	}
	return 0
}

func (x *ProtoInfo) GetFaults() uint64 {
	if x != nil {
		return x.Faults
	}
	return 0
}

func (x *ProtoInfo) GetBlockSize() uint64 {
	if x != nil {
		return x.BlockSize
	}
	return 0
}

type SyncHSConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    uint64     `protobuf:"varint,1,opt,name=Id,proto3" json:"Id,omitempty"` // Replica ID
	Info  *ProtoInfo `protobuf:"bytes,2,opt,name=Info,proto3" json:"Info,omitempty"`
	Delta float64    `protobuf:"fixed64,3,opt,name=Delta,proto3" json:"Delta,omitempty"` // In seconds
}

func (x *SyncHSConfig) Reset() {
	*x = SyncHSConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_protocol_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncHSConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncHSConfig) ProtoMessage() {}

func (x *SyncHSConfig) ProtoReflect() protoreflect.Message {
	mi := &file_config_protocol_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncHSConfig.ProtoReflect.Descriptor instead.
func (*SyncHSConfig) Descriptor() ([]byte, []int) {
	return file_config_protocol_proto_rawDescGZIP(), []int{1}
}

func (x *SyncHSConfig) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SyncHSConfig) GetInfo() *ProtoInfo {
	if x != nil {
		return x.Info
	}
	return nil
}

func (x *SyncHSConfig) GetDelta() float64 {
	if x != nil {
		return x.Delta
	}
	return 0
}

var File_config_protocol_proto protoreflect.FileDescriptor

var file_config_protocol_proto_rawDesc = []byte{
	0x0a, 0x15, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22,
	0x5d, 0x0a, 0x09, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08,
	0x4e, 0x6f, 0x64, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08,
	0x4e, 0x6f, 0x64, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x46, 0x61, 0x75, 0x6c,
	0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x46, 0x61, 0x75, 0x6c, 0x74, 0x73,
	0x12, 0x1c, 0x0a, 0x09, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x09, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x5b,
	0x0a, 0x0c, 0x53, 0x79, 0x6e, 0x63, 0x48, 0x53, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x0e,
	0x0a, 0x02, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x49, 0x64, 0x12, 0x25,
	0x0a, 0x04, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x04, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x44, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x44, 0x65, 0x6c, 0x74, 0x61, 0x42, 0x09, 0x5a, 0x07, 0x2f,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_config_protocol_proto_rawDescOnce sync.Once
	file_config_protocol_proto_rawDescData = file_config_protocol_proto_rawDesc
)

func file_config_protocol_proto_rawDescGZIP() []byte {
	file_config_protocol_proto_rawDescOnce.Do(func() {
		file_config_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_protocol_proto_rawDescData)
	})
	return file_config_protocol_proto_rawDescData
}

var file_config_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_config_protocol_proto_goTypes = []interface{}{
	(*ProtoInfo)(nil),    // 0: config.ProtoInfo
	(*SyncHSConfig)(nil), // 1: config.SyncHSConfig
}
var file_config_protocol_proto_depIdxs = []int32{
	0, // 0: config.SyncHSConfig.Info:type_name -> config.ProtoInfo
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_config_protocol_proto_init() }
func file_config_protocol_proto_init() {
	if File_config_protocol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_config_protocol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_config_protocol_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncHSConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_config_protocol_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_protocol_proto_goTypes,
		DependencyIndexes: file_config_protocol_proto_depIdxs,
		MessageInfos:      file_config_protocol_proto_msgTypes,
	}.Build()
	File_config_protocol_proto = out.File
	file_config_protocol_proto_rawDesc = nil
	file_config_protocol_proto_goTypes = nil
	file_config_protocol_proto_depIdxs = nil
}