package client

import (
	"encoding/binary"
	"testing"

	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
)

func TestCodecUsesConfirmedNetworkByteOrderAndHeaders(t *testing.T) {
	inbound := make([]byte, 8)
	binary.BigEndian.PutUint16(inbound[:2], uint16(commonpb.MessageID_LoginPlayerReq))
	binary.BigEndian.PutUint32(inbound[2:6], 7)
	copy(inbound[6:], []byte{1, 2})
	decoded, err := decodeRequest(inbound)
	if err != nil || decoded.messageID != commonpb.MessageID_LoginPlayerReq || decoded.sequence != 7 || len(decoded.body) != 2 {
		t.Fatalf("decodeRequest() = %+v, %v", decoded, err)
	}

	encoded, err := encodeClientMessage(rpcapi.ClientMessage{
		MessageID: commonpb.MessageID_LoginPlayerRes, Sequence: 7,
		ErrorCode: commonpb.ErrorCode_ERROR_CODE_GATEWAY_TOKEN_INVALID, Body: []byte{3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 11 || binary.BigEndian.Uint32(encoded[2:6]) != 7 ||
		int32(binary.BigEndian.Uint32(encoded[6:10])) != int32(commonpb.ErrorCode_ERROR_CODE_GATEWAY_TOKEN_INVALID) {
		t.Fatalf("响应头编码错误: %v", encoded)
	}
}

func TestDecodeRejectsClientPushSequence(t *testing.T) {
	if _, err := decodeRequest(make([]byte, requestHeaderSize)); err == nil {
		t.Fatal("客户端 Sequence=0 必须拒绝")
	}
}
