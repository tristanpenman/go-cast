// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: internal/cast/cast_channel.proto

package channel

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

type SignatureAlgorithm int32

const (
	SignatureAlgorithm_UNSPECIFIED     SignatureAlgorithm = 0
	SignatureAlgorithm_RSASSA_PKCS1v15 SignatureAlgorithm = 1
	SignatureAlgorithm_RSASSA_PSS      SignatureAlgorithm = 2
)

// Enum value maps for SignatureAlgorithm.
var (
	SignatureAlgorithm_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "RSASSA_PKCS1v15",
		2: "RSASSA_PSS",
	}
	SignatureAlgorithm_value = map[string]int32{
		"UNSPECIFIED":     0,
		"RSASSA_PKCS1v15": 1,
		"RSASSA_PSS":      2,
	}
)

func (x SignatureAlgorithm) Enum() *SignatureAlgorithm {
	p := new(SignatureAlgorithm)
	*p = x
	return p
}

func (x SignatureAlgorithm) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SignatureAlgorithm) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_cast_cast_channel_proto_enumTypes[0].Descriptor()
}

func (SignatureAlgorithm) Type() protoreflect.EnumType {
	return &file_internal_cast_cast_channel_proto_enumTypes[0]
}

func (x SignatureAlgorithm) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *SignatureAlgorithm) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = SignatureAlgorithm(num)
	return nil
}

// Deprecated: Use SignatureAlgorithm.Descriptor instead.
func (SignatureAlgorithm) EnumDescriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{0}
}

type HashAlgorithm int32

const (
	HashAlgorithm_SHA1   HashAlgorithm = 0
	HashAlgorithm_SHA256 HashAlgorithm = 1
)

// Enum value maps for HashAlgorithm.
var (
	HashAlgorithm_name = map[int32]string{
		0: "SHA1",
		1: "SHA256",
	}
	HashAlgorithm_value = map[string]int32{
		"SHA1":   0,
		"SHA256": 1,
	}
)

func (x HashAlgorithm) Enum() *HashAlgorithm {
	p := new(HashAlgorithm)
	*p = x
	return p
}

func (x HashAlgorithm) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HashAlgorithm) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_cast_cast_channel_proto_enumTypes[1].Descriptor()
}

func (HashAlgorithm) Type() protoreflect.EnumType {
	return &file_internal_cast_cast_channel_proto_enumTypes[1]
}

func (x HashAlgorithm) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *HashAlgorithm) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = HashAlgorithm(num)
	return nil
}

// Deprecated: Use HashAlgorithm.Descriptor instead.
func (HashAlgorithm) EnumDescriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{1}
}

// Always pass a version of the protocol for future compatibility
// requirements.
type CastMessage_ProtocolVersion int32

const (
	CastMessage_CASTV2_1_0 CastMessage_ProtocolVersion = 0
)

// Enum value maps for CastMessage_ProtocolVersion.
var (
	CastMessage_ProtocolVersion_name = map[int32]string{
		0: "CASTV2_1_0",
	}
	CastMessage_ProtocolVersion_value = map[string]int32{
		"CASTV2_1_0": 0,
	}
)

func (x CastMessage_ProtocolVersion) Enum() *CastMessage_ProtocolVersion {
	p := new(CastMessage_ProtocolVersion)
	*p = x
	return p
}

func (x CastMessage_ProtocolVersion) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CastMessage_ProtocolVersion) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_cast_cast_channel_proto_enumTypes[2].Descriptor()
}

func (CastMessage_ProtocolVersion) Type() protoreflect.EnumType {
	return &file_internal_cast_cast_channel_proto_enumTypes[2]
}

func (x CastMessage_ProtocolVersion) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *CastMessage_ProtocolVersion) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = CastMessage_ProtocolVersion(num)
	return nil
}

// Deprecated: Use CastMessage_ProtocolVersion.Descriptor instead.
func (CastMessage_ProtocolVersion) EnumDescriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{0, 0}
}

// What type of data do we have in this message.
type CastMessage_PayloadType int32

const (
	CastMessage_STRING CastMessage_PayloadType = 0
	CastMessage_BINARY CastMessage_PayloadType = 1
)

// Enum value maps for CastMessage_PayloadType.
var (
	CastMessage_PayloadType_name = map[int32]string{
		0: "STRING",
		1: "BINARY",
	}
	CastMessage_PayloadType_value = map[string]int32{
		"STRING": 0,
		"BINARY": 1,
	}
)

func (x CastMessage_PayloadType) Enum() *CastMessage_PayloadType {
	p := new(CastMessage_PayloadType)
	*p = x
	return p
}

func (x CastMessage_PayloadType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CastMessage_PayloadType) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_cast_cast_channel_proto_enumTypes[3].Descriptor()
}

func (CastMessage_PayloadType) Type() protoreflect.EnumType {
	return &file_internal_cast_cast_channel_proto_enumTypes[3]
}

func (x CastMessage_PayloadType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *CastMessage_PayloadType) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = CastMessage_PayloadType(num)
	return nil
}

// Deprecated: Use CastMessage_PayloadType.Descriptor instead.
func (CastMessage_PayloadType) EnumDescriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{0, 1}
}

type AuthError_ErrorType int32

const (
	AuthError_INTERNAL_ERROR                  AuthError_ErrorType = 0
	AuthError_NO_TLS                          AuthError_ErrorType = 1 // The underlying connection is not TLS
	AuthError_SIGNATURE_ALGORITHM_UNAVAILABLE AuthError_ErrorType = 2
)

// Enum value maps for AuthError_ErrorType.
var (
	AuthError_ErrorType_name = map[int32]string{
		0: "INTERNAL_ERROR",
		1: "NO_TLS",
		2: "SIGNATURE_ALGORITHM_UNAVAILABLE",
	}
	AuthError_ErrorType_value = map[string]int32{
		"INTERNAL_ERROR":                  0,
		"NO_TLS":                          1,
		"SIGNATURE_ALGORITHM_UNAVAILABLE": 2,
	}
)

func (x AuthError_ErrorType) Enum() *AuthError_ErrorType {
	p := new(AuthError_ErrorType)
	*p = x
	return p
}

func (x AuthError_ErrorType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuthError_ErrorType) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_cast_cast_channel_proto_enumTypes[4].Descriptor()
}

func (AuthError_ErrorType) Type() protoreflect.EnumType {
	return &file_internal_cast_cast_channel_proto_enumTypes[4]
}

func (x AuthError_ErrorType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *AuthError_ErrorType) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = AuthError_ErrorType(num)
	return nil
}

// Deprecated: Use AuthError_ErrorType.Descriptor instead.
func (AuthError_ErrorType) EnumDescriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{3, 0}
}

type CastMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProtocolVersion *CastMessage_ProtocolVersion `protobuf:"varint,1,req,name=protocol_version,json=protocolVersion,enum=cast_channel.CastMessage_ProtocolVersion" json:"protocol_version,omitempty"`
	// source and destination ids identify the origin and destination of the
	// message.  They are used to route messages between endpoints that share a
	// device-to-device channel.
	//
	// For messages between applications:
	//   - The sender application id is a unique identifier generated on behalf of
	//     the sender application.
	//   - The receiver id is always the the session id for the application.
	//
	// For messages to or from the sender or receiver platform, the special ids
	// 'sender-0' and 'receiver-0' can be used.
	//
	// For messages intended for all endpoints using a given channel, the
	// wildcard destination_id '*' can be used.
	SourceId      *string `protobuf:"bytes,2,req,name=source_id,json=sourceId" json:"source_id,omitempty"`
	DestinationId *string `protobuf:"bytes,3,req,name=destination_id,json=destinationId" json:"destination_id,omitempty"`
	// This is the core multiplexing key.  All messages are sent on a namespace
	// and endpoints sharing a channel listen on one or more namespaces.  The
	// namespace defines the protocol and semantics of the message.
	Namespace   *string                  `protobuf:"bytes,4,req,name=namespace" json:"namespace,omitempty"`
	PayloadType *CastMessage_PayloadType `protobuf:"varint,5,req,name=payload_type,json=payloadType,enum=cast_channel.CastMessage_PayloadType" json:"payload_type,omitempty"`
	// Depending on payload_type, exactly one of the following optional fields
	// will always be set.
	PayloadUtf8   *string `protobuf:"bytes,6,opt,name=payload_utf8,json=payloadUtf8" json:"payload_utf8,omitempty"`
	PayloadBinary []byte  `protobuf:"bytes,7,opt,name=payload_binary,json=payloadBinary" json:"payload_binary,omitempty"`
}

func (x *CastMessage) Reset() {
	*x = CastMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_cast_cast_channel_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CastMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CastMessage) ProtoMessage() {}

func (x *CastMessage) ProtoReflect() protoreflect.Message {
	mi := &file_internal_cast_cast_channel_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CastMessage.ProtoReflect.Descriptor instead.
func (*CastMessage) Descriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{0}
}

func (x *CastMessage) GetProtocolVersion() CastMessage_ProtocolVersion {
	if x != nil && x.ProtocolVersion != nil {
		return *x.ProtocolVersion
	}
	return CastMessage_CASTV2_1_0
}

func (x *CastMessage) GetSourceId() string {
	if x != nil && x.SourceId != nil {
		return *x.SourceId
	}
	return ""
}

func (x *CastMessage) GetDestinationId() string {
	if x != nil && x.DestinationId != nil {
		return *x.DestinationId
	}
	return ""
}

func (x *CastMessage) GetNamespace() string {
	if x != nil && x.Namespace != nil {
		return *x.Namespace
	}
	return ""
}

func (x *CastMessage) GetPayloadType() CastMessage_PayloadType {
	if x != nil && x.PayloadType != nil {
		return *x.PayloadType
	}
	return CastMessage_STRING
}

func (x *CastMessage) GetPayloadUtf8() string {
	if x != nil && x.PayloadUtf8 != nil {
		return *x.PayloadUtf8
	}
	return ""
}

func (x *CastMessage) GetPayloadBinary() []byte {
	if x != nil {
		return x.PayloadBinary
	}
	return nil
}

// Messages for authentication protocol between a sender and a receiver.
type AuthChallenge struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SignatureAlgorithm *SignatureAlgorithm `protobuf:"varint,1,opt,name=signature_algorithm,json=signatureAlgorithm,enum=cast_channel.SignatureAlgorithm,def=1" json:"signature_algorithm,omitempty"`
	SenderNonce        []byte              `protobuf:"bytes,2,opt,name=sender_nonce,json=senderNonce" json:"sender_nonce,omitempty"`
	HashAlgorithm      *HashAlgorithm      `protobuf:"varint,3,opt,name=hash_algorithm,json=hashAlgorithm,enum=cast_channel.HashAlgorithm,def=0" json:"hash_algorithm,omitempty"`
}

// Default values for AuthChallenge fields.
const (
	Default_AuthChallenge_SignatureAlgorithm = SignatureAlgorithm_RSASSA_PKCS1v15
	Default_AuthChallenge_HashAlgorithm      = HashAlgorithm_SHA1
)

func (x *AuthChallenge) Reset() {
	*x = AuthChallenge{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_cast_cast_channel_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthChallenge) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthChallenge) ProtoMessage() {}

func (x *AuthChallenge) ProtoReflect() protoreflect.Message {
	mi := &file_internal_cast_cast_channel_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthChallenge.ProtoReflect.Descriptor instead.
func (*AuthChallenge) Descriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{1}
}

func (x *AuthChallenge) GetSignatureAlgorithm() SignatureAlgorithm {
	if x != nil && x.SignatureAlgorithm != nil {
		return *x.SignatureAlgorithm
	}
	return Default_AuthChallenge_SignatureAlgorithm
}

func (x *AuthChallenge) GetSenderNonce() []byte {
	if x != nil {
		return x.SenderNonce
	}
	return nil
}

func (x *AuthChallenge) GetHashAlgorithm() HashAlgorithm {
	if x != nil && x.HashAlgorithm != nil {
		return *x.HashAlgorithm
	}
	return Default_AuthChallenge_HashAlgorithm
}

type AuthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Signature               []byte              `protobuf:"bytes,1,req,name=signature" json:"signature,omitempty"`
	ClientAuthCertificate   []byte              `protobuf:"bytes,2,req,name=client_auth_certificate,json=clientAuthCertificate" json:"client_auth_certificate,omitempty"`
	IntermediateCertificate [][]byte            `protobuf:"bytes,3,rep,name=intermediate_certificate,json=intermediateCertificate" json:"intermediate_certificate,omitempty"`
	SignatureAlgorithm      *SignatureAlgorithm `protobuf:"varint,4,opt,name=signature_algorithm,json=signatureAlgorithm,enum=cast_channel.SignatureAlgorithm,def=1" json:"signature_algorithm,omitempty"`
	SenderNonce             []byte              `protobuf:"bytes,5,opt,name=sender_nonce,json=senderNonce" json:"sender_nonce,omitempty"`
	HashAlgorithm           *HashAlgorithm      `protobuf:"varint,6,opt,name=hash_algorithm,json=hashAlgorithm,enum=cast_channel.HashAlgorithm,def=0" json:"hash_algorithm,omitempty"`
	Crl                     []byte              `protobuf:"bytes,7,opt,name=crl" json:"crl,omitempty"`
}

// Default values for AuthResponse fields.
const (
	Default_AuthResponse_SignatureAlgorithm = SignatureAlgorithm_RSASSA_PKCS1v15
	Default_AuthResponse_HashAlgorithm      = HashAlgorithm_SHA1
)

func (x *AuthResponse) Reset() {
	*x = AuthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_cast_cast_channel_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthResponse) ProtoMessage() {}

func (x *AuthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_cast_cast_channel_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthResponse.ProtoReflect.Descriptor instead.
func (*AuthResponse) Descriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{2}
}

func (x *AuthResponse) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *AuthResponse) GetClientAuthCertificate() []byte {
	if x != nil {
		return x.ClientAuthCertificate
	}
	return nil
}

func (x *AuthResponse) GetIntermediateCertificate() [][]byte {
	if x != nil {
		return x.IntermediateCertificate
	}
	return nil
}

func (x *AuthResponse) GetSignatureAlgorithm() SignatureAlgorithm {
	if x != nil && x.SignatureAlgorithm != nil {
		return *x.SignatureAlgorithm
	}
	return Default_AuthResponse_SignatureAlgorithm
}

func (x *AuthResponse) GetSenderNonce() []byte {
	if x != nil {
		return x.SenderNonce
	}
	return nil
}

func (x *AuthResponse) GetHashAlgorithm() HashAlgorithm {
	if x != nil && x.HashAlgorithm != nil {
		return *x.HashAlgorithm
	}
	return Default_AuthResponse_HashAlgorithm
}

func (x *AuthResponse) GetCrl() []byte {
	if x != nil {
		return x.Crl
	}
	return nil
}

type AuthError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrorType *AuthError_ErrorType `protobuf:"varint,1,req,name=error_type,json=errorType,enum=cast_channel.AuthError_ErrorType" json:"error_type,omitempty"`
}

func (x *AuthError) Reset() {
	*x = AuthError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_cast_cast_channel_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthError) ProtoMessage() {}

func (x *AuthError) ProtoReflect() protoreflect.Message {
	mi := &file_internal_cast_cast_channel_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthError.ProtoReflect.Descriptor instead.
func (*AuthError) Descriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{3}
}

func (x *AuthError) GetErrorType() AuthError_ErrorType {
	if x != nil && x.ErrorType != nil {
		return *x.ErrorType
	}
	return AuthError_INTERNAL_ERROR
}

type DeviceAuthMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Request fields
	Challenge *AuthChallenge `protobuf:"bytes,1,opt,name=challenge" json:"challenge,omitempty"`
	// Response fields
	Response *AuthResponse `protobuf:"bytes,2,opt,name=response" json:"response,omitempty"`
	Error    *AuthError    `protobuf:"bytes,3,opt,name=error" json:"error,omitempty"`
}

func (x *DeviceAuthMessage) Reset() {
	*x = DeviceAuthMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_cast_cast_channel_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceAuthMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceAuthMessage) ProtoMessage() {}

func (x *DeviceAuthMessage) ProtoReflect() protoreflect.Message {
	mi := &file_internal_cast_cast_channel_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceAuthMessage.ProtoReflect.Descriptor instead.
func (*DeviceAuthMessage) Descriptor() ([]byte, []int) {
	return file_internal_cast_cast_channel_proto_rawDescGZIP(), []int{4}
}

func (x *DeviceAuthMessage) GetChallenge() *AuthChallenge {
	if x != nil {
		return x.Challenge
	}
	return nil
}

func (x *DeviceAuthMessage) GetResponse() *AuthResponse {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *DeviceAuthMessage) GetError() *AuthError {
	if x != nil {
		return x.Error
	}
	return nil
}

var File_internal_cast_cast_channel_proto protoreflect.FileDescriptor

var file_internal_cast_cast_channel_proto_rawDesc = []byte{
	0x0a, 0x20, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x63, 0x61, 0x73, 0x74, 0x2f,
	0x63, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0c, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c,
	0x22, 0xa3, 0x03, 0x0a, 0x0b, 0x43, 0x61, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x54, 0x0a, 0x10, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0e, 0x32, 0x29, 0x2e, 0x63, 0x61, 0x73,
	0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x43, 0x61, 0x73, 0x74, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x02, 0x28, 0x09, 0x52, 0x08, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x02, 0x28, 0x09, 0x52, 0x0d, 0x64, 0x65, 0x73,
	0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x04, 0x20, 0x02, 0x28, 0x09, 0x52, 0x09, 0x6e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x48, 0x0a, 0x0c, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x02, 0x28, 0x0e, 0x32, 0x25,
	0x2e, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x43, 0x61,
	0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0b, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x75, 0x74,
	0x66, 0x38, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x55, 0x74, 0x66, 0x38, 0x12, 0x25, 0x0a, 0x0e, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x5f, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x70,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x22, 0x21, 0x0a, 0x0f,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x0e, 0x0a, 0x0a, 0x43, 0x41, 0x53, 0x54, 0x56, 0x32, 0x5f, 0x31, 0x5f, 0x30, 0x10, 0x00, 0x22,
	0x25, 0x0a, 0x0b, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0a,
	0x0a, 0x06, 0x53, 0x54, 0x52, 0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x42, 0x49,
	0x4e, 0x41, 0x52, 0x59, 0x10, 0x01, 0x22, 0xe0, 0x01, 0x0a, 0x0d, 0x41, 0x75, 0x74, 0x68, 0x43,
	0x68, 0x61, 0x6c, 0x6c, 0x65, 0x6e, 0x67, 0x65, 0x12, 0x62, 0x0a, 0x13, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61,
	0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x41, 0x6c,
	0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x3a, 0x0f, 0x52, 0x53, 0x41, 0x53, 0x53, 0x41, 0x5f,
	0x50, 0x4b, 0x43, 0x53, 0x31, 0x76, 0x31, 0x35, 0x52, 0x12, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x21, 0x0a, 0x0c,
	0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0b, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x4e, 0x6f, 0x6e, 0x63, 0x65, 0x12,
	0x48, 0x0a, 0x0e, 0x68, 0x61, 0x73, 0x68, 0x5f, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68,
	0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x63,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x41, 0x6c, 0x67, 0x6f, 0x72,
	0x69, 0x74, 0x68, 0x6d, 0x3a, 0x04, 0x53, 0x48, 0x41, 0x31, 0x52, 0x0d, 0x68, 0x61, 0x73, 0x68,
	0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x22, 0x82, 0x03, 0x0a, 0x0c, 0x41, 0x75,
	0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0c, 0x52, 0x09, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x36, 0x0a, 0x17, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0c, 0x52, 0x15, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x41, 0x75, 0x74, 0x68, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65,
	0x12, 0x39, 0x0a, 0x18, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x74, 0x65,
	0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0c, 0x52, 0x17, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x74, 0x65,
	0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x62, 0x0a, 0x13, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74,
	0x68, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x5f,
	0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x3a, 0x0f, 0x52, 0x53, 0x41, 0x53,
	0x53, 0x41, 0x5f, 0x50, 0x4b, 0x43, 0x53, 0x31, 0x76, 0x31, 0x35, 0x52, 0x12, 0x73, 0x69, 0x67,
	0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12,
	0x21, 0x0a, 0x0c, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x4e, 0x6f, 0x6e,
	0x63, 0x65, 0x12, 0x48, 0x0a, 0x0e, 0x68, 0x61, 0x73, 0x68, 0x5f, 0x61, 0x6c, 0x67, 0x6f, 0x72,
	0x69, 0x74, 0x68, 0x6d, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x63, 0x61, 0x73,
	0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x41, 0x6c,
	0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x3a, 0x04, 0x53, 0x48, 0x41, 0x31, 0x52, 0x0d, 0x68,
	0x61, 0x73, 0x68, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x10, 0x0a, 0x03,
	0x63, 0x72, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x63, 0x72, 0x6c, 0x22, 0x9f,
	0x01, 0x0a, 0x09, 0x41, 0x75, 0x74, 0x68, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x40, 0x0a, 0x0a,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0e,
	0x32, 0x21, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e,
	0x41, 0x75, 0x74, 0x68, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x09, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x54, 0x79, 0x70, 0x65, 0x22, 0x50,
	0x0a, 0x09, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x0e, 0x49,
	0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x00, 0x12,
	0x0a, 0x0a, 0x06, 0x4e, 0x4f, 0x5f, 0x54, 0x4c, 0x53, 0x10, 0x01, 0x12, 0x23, 0x0a, 0x1f, 0x53,
	0x49, 0x47, 0x4e, 0x41, 0x54, 0x55, 0x52, 0x45, 0x5f, 0x41, 0x4c, 0x47, 0x4f, 0x52, 0x49, 0x54,
	0x48, 0x4d, 0x5f, 0x55, 0x4e, 0x41, 0x56, 0x41, 0x49, 0x4c, 0x41, 0x42, 0x4c, 0x45, 0x10, 0x02,
	0x22, 0xb5, 0x01, 0x0a, 0x11, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x41, 0x75, 0x74, 0x68, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x39, 0x0a, 0x09, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x65,
	0x6e, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x61, 0x73, 0x74,
	0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x43, 0x68, 0x61,
	0x6c, 0x6c, 0x65, 0x6e, 0x67, 0x65, 0x52, 0x09, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x65, 0x6e, 0x67,
	0x65, 0x12, 0x36, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52,
	0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d, 0x0a, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x5f,
	0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2a, 0x4a, 0x0a, 0x12, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x0f,
	0x0a, 0x0b, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x13, 0x0a, 0x0f, 0x52, 0x53, 0x41, 0x53, 0x53, 0x41, 0x5f, 0x50, 0x4b, 0x43, 0x53, 0x31, 0x76,
	0x31, 0x35, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x52, 0x53, 0x41, 0x53, 0x53, 0x41, 0x5f, 0x50,
	0x53, 0x53, 0x10, 0x02, 0x2a, 0x25, 0x0a, 0x0d, 0x48, 0x61, 0x73, 0x68, 0x41, 0x6c, 0x67, 0x6f,
	0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x08, 0x0a, 0x04, 0x53, 0x48, 0x41, 0x31, 0x10, 0x00, 0x12,
	0x0a, 0x0a, 0x06, 0x53, 0x48, 0x41, 0x32, 0x35, 0x36, 0x10, 0x01, 0x42, 0x35, 0x48, 0x03, 0x5a,
	0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x72, 0x69, 0x73,
	0x74, 0x61, 0x6e, 0x70, 0x65, 0x6e, 0x6d, 0x61, 0x6e, 0x2f, 0x67, 0x6f, 0x2d, 0x63, 0x61, 0x73,
	0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x63, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c,
}

var (
	file_internal_cast_cast_channel_proto_rawDescOnce sync.Once
	file_internal_cast_cast_channel_proto_rawDescData = file_internal_cast_cast_channel_proto_rawDesc
)

func file_internal_cast_cast_channel_proto_rawDescGZIP() []byte {
	file_internal_cast_cast_channel_proto_rawDescOnce.Do(func() {
		file_internal_cast_cast_channel_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_cast_cast_channel_proto_rawDescData)
	})
	return file_internal_cast_cast_channel_proto_rawDescData
}

var file_internal_cast_cast_channel_proto_enumTypes = make([]protoimpl.EnumInfo, 5)
var file_internal_cast_cast_channel_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_internal_cast_cast_channel_proto_goTypes = []interface{}{
	(SignatureAlgorithm)(0),          // 0: cast_channel.SignatureAlgorithm
	(HashAlgorithm)(0),               // 1: cast_channel.HashAlgorithm
	(CastMessage_ProtocolVersion)(0), // 2: cast_channel.CastMessage.ProtocolVersion
	(CastMessage_PayloadType)(0),     // 3: cast_channel.CastMessage.PayloadType
	(AuthError_ErrorType)(0),         // 4: cast_channel.AuthError.ErrorType
	(*CastMessage)(nil),              // 5: cast_channel.CastMessage
	(*AuthChallenge)(nil),            // 6: cast_channel.AuthChallenge
	(*AuthResponse)(nil),             // 7: cast_channel.AuthResponse
	(*AuthError)(nil),                // 8: cast_channel.AuthError
	(*DeviceAuthMessage)(nil),        // 9: cast_channel.DeviceAuthMessage
}
var file_internal_cast_cast_channel_proto_depIdxs = []int32{
	2,  // 0: cast_channel.CastMessage.protocol_version:type_name -> cast_channel.CastMessage.ProtocolVersion
	3,  // 1: cast_channel.CastMessage.payload_type:type_name -> cast_channel.CastMessage.PayloadType
	0,  // 2: cast_channel.AuthChallenge.signature_algorithm:type_name -> cast_channel.SignatureAlgorithm
	1,  // 3: cast_channel.AuthChallenge.hash_algorithm:type_name -> cast_channel.HashAlgorithm
	0,  // 4: cast_channel.AuthResponse.signature_algorithm:type_name -> cast_channel.SignatureAlgorithm
	1,  // 5: cast_channel.AuthResponse.hash_algorithm:type_name -> cast_channel.HashAlgorithm
	4,  // 6: cast_channel.AuthError.error_type:type_name -> cast_channel.AuthError.ErrorType
	6,  // 7: cast_channel.DeviceAuthMessage.challenge:type_name -> cast_channel.AuthChallenge
	7,  // 8: cast_channel.DeviceAuthMessage.response:type_name -> cast_channel.AuthResponse
	8,  // 9: cast_channel.DeviceAuthMessage.error:type_name -> cast_channel.AuthError
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_internal_cast_cast_channel_proto_init() }
func file_internal_cast_cast_channel_proto_init() {
	if File_internal_cast_cast_channel_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_cast_cast_channel_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CastMessage); i {
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
		file_internal_cast_cast_channel_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthChallenge); i {
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
		file_internal_cast_cast_channel_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthResponse); i {
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
		file_internal_cast_cast_channel_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthError); i {
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
		file_internal_cast_cast_channel_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceAuthMessage); i {
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
			RawDescriptor: file_internal_cast_cast_channel_proto_rawDesc,
			NumEnums:      5,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_cast_cast_channel_proto_goTypes,
		DependencyIndexes: file_internal_cast_cast_channel_proto_depIdxs,
		EnumInfos:         file_internal_cast_cast_channel_proto_enumTypes,
		MessageInfos:      file_internal_cast_cast_channel_proto_msgTypes,
	}.Build()
	File_internal_cast_cast_channel_proto = out.File
	file_internal_cast_cast_channel_proto_rawDesc = nil
	file_internal_cast_cast_channel_proto_goTypes = nil
	file_internal_cast_cast_channel_proto_depIdxs = nil
}