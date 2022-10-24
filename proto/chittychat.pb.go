// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.6
// source: proto/chittychat.proto

package proto

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

type Information struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientName string `protobuf:"bytes,1,opt,name=clientName,proto3" json:"clientName,omitempty"`
}

func (x *Information) Reset() {
	*x = Information{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_chittychat_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Information) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Information) ProtoMessage() {}

func (x *Information) ProtoReflect() protoreflect.Message {
	mi := &file_proto_chittychat_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Information.ProtoReflect.Descriptor instead.
func (*Information) Descriptor() ([]byte, []int) {
	return file_proto_chittychat_proto_rawDescGZIP(), []int{0}
}

func (x *Information) GetClientName() string {
	if x != nil {
		return x.ClientName
	}
	return ""
}

// notice the stream of StatusChange, this is since other people joining chatroom
// will be notified through this stream aswell
type StatusChange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// if joined is false on first message in stream, then you are not registered - probably due
	// to non-unique name. for the rest of the stream however, it's a way of determining whether
	// a person left or joined for broadcasting that information.
	Joined bool `protobuf:"varint,1,opt,name=joined,proto3" json:"joined,omitempty"`
	// name of client that joined
	ClientName string `protobuf:"bytes,2,opt,name=clientName,proto3" json:"clientName,omitempty"`
	// id of client, this is used in Message, to not send the full clientName each time
	// clients are expected to keep track of mapping id -> clientName
	Id      int32 `protobuf:"varint,3,opt,name=id,proto3" json:"id,omitempty"`
	Lamport int64 `protobuf:"varint,4,opt,name=lamport,proto3" json:"lamport,omitempty"`
}

func (x *StatusChange) Reset() {
	*x = StatusChange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_chittychat_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusChange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusChange) ProtoMessage() {}

func (x *StatusChange) ProtoReflect() protoreflect.Message {
	mi := &file_proto_chittychat_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusChange.ProtoReflect.Descriptor instead.
func (*StatusChange) Descriptor() ([]byte, []int) {
	return file_proto_chittychat_proto_rawDescGZIP(), []int{1}
}

func (x *StatusChange) GetJoined() bool {
	if x != nil {
		return x.Joined
	}
	return false
}

func (x *StatusChange) GetClientName() string {
	if x != nil {
		return x.ClientName
	}
	return ""
}

func (x *StatusChange) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *StatusChange) GetLamport() int64 {
	if x != nil {
		return x.Lamport
	}
	return 0
}

type MessageSent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token   string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Lamport int64  `protobuf:"varint,3,opt,name=lamport,proto3" json:"lamport,omitempty"`
}

func (x *MessageSent) Reset() {
	*x = MessageSent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_chittychat_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageSent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageSent) ProtoMessage() {}

func (x *MessageSent) ProtoReflect() protoreflect.Message {
	mi := &file_proto_chittychat_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageSent.ProtoReflect.Descriptor instead.
func (*MessageSent) Descriptor() ([]byte, []int) {
	return file_proto_chittychat_proto_rawDescGZIP(), []int{2}
}

func (x *MessageSent) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *MessageSent) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *MessageSent) GetLamport() int64 {
	if x != nil {
		return x.Lamport
	}
	return 0
}

type MessageRecv struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Lamport int64  `protobuf:"varint,3,opt,name=lamport,proto3" json:"lamport,omitempty"`
}

func (x *MessageRecv) Reset() {
	*x = MessageRecv{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_chittychat_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageRecv) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageRecv) ProtoMessage() {}

func (x *MessageRecv) ProtoReflect() protoreflect.Message {
	mi := &file_proto_chittychat_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageRecv.ProtoReflect.Descriptor instead.
func (*MessageRecv) Descriptor() ([]byte, []int) {
	return file_proto_chittychat_proto_rawDescGZIP(), []int{3}
}

func (x *MessageRecv) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *MessageRecv) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *MessageRecv) GetLamport() int64 {
	if x != nil {
		return x.Lamport
	}
	return 0
}

var File_proto_chittychat_proto protoreflect.FileDescriptor

var file_proto_chittychat_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x68, 0x69, 0x74, 0x74, 0x79, 0x63, 0x68,
	0x61, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x2d, 0x0a, 0x0b, 0x49, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1e,
	0x0a, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x70,
	0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x6a, 0x6f, 0x69, 0x6e, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06,
	0x6a, 0x6f, 0x69, 0x6e, 0x65, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72,
	0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74,
	0x22, 0x57, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x53, 0x65, 0x6e, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x51, 0x0a, 0x0b, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x63, 0x76, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x07, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x32, 0x76, 0x0a, 0x0a,
	0x43, 0x68, 0x69, 0x74, 0x74, 0x79, 0x43, 0x68, 0x61, 0x74, 0x12, 0x31, 0x0a, 0x04, 0x4a, 0x6f,
	0x69, 0x6e, 0x12, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x13, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x30, 0x01, 0x12, 0x35, 0x0a,
	0x07, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x12, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x53, 0x65, 0x6e, 0x74, 0x1a, 0x12, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x63, 0x76,
	0x28, 0x01, 0x30, 0x01, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x65, 0x64, 0x49, 0x6e, 0x53, 0x70, 0x61, 0x63,
	0x65, 0x2f, 0x43, 0x68, 0x69, 0x74, 0x74, 0x79, 0x43, 0x68, 0x61, 0x74, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_chittychat_proto_rawDescOnce sync.Once
	file_proto_chittychat_proto_rawDescData = file_proto_chittychat_proto_rawDesc
)

func file_proto_chittychat_proto_rawDescGZIP() []byte {
	file_proto_chittychat_proto_rawDescOnce.Do(func() {
		file_proto_chittychat_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_chittychat_proto_rawDescData)
	})
	return file_proto_chittychat_proto_rawDescData
}

var file_proto_chittychat_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_chittychat_proto_goTypes = []interface{}{
	(*Information)(nil),  // 0: proto.Information
	(*StatusChange)(nil), // 1: proto.StatusChange
	(*MessageSent)(nil),  // 2: proto.MessageSent
	(*MessageRecv)(nil),  // 3: proto.MessageRecv
}
var file_proto_chittychat_proto_depIdxs = []int32{
	0, // 0: proto.ChittyChat.Join:input_type -> proto.Information
	2, // 1: proto.ChittyChat.Publish:input_type -> proto.MessageSent
	1, // 2: proto.ChittyChat.Join:output_type -> proto.StatusChange
	3, // 3: proto.ChittyChat.Publish:output_type -> proto.MessageRecv
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_chittychat_proto_init() }
func file_proto_chittychat_proto_init() {
	if File_proto_chittychat_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_chittychat_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Information); i {
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
		file_proto_chittychat_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusChange); i {
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
		file_proto_chittychat_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageSent); i {
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
		file_proto_chittychat_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageRecv); i {
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
			RawDescriptor: file_proto_chittychat_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_chittychat_proto_goTypes,
		DependencyIndexes: file_proto_chittychat_proto_depIdxs,
		MessageInfos:      file_proto_chittychat_proto_msgTypes,
	}.Build()
	File_proto_chittychat_proto = out.File
	file_proto_chittychat_proto_rawDesc = nil
	file_proto_chittychat_proto_goTypes = nil
	file_proto_chittychat_proto_depIdxs = nil
}
