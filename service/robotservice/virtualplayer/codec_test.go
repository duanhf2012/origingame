package virtualplayer

import (
	"encoding/binary"
	"testing"

	commonpb "origingame/protocol/common"
)

func TestCodecMatchesGatewayHeaders(t *testing.T) {
	request, err := EncodeRequest(commonpb.MessageID_LoginPlayerReq, 7, []byte{1, 2})
	if err != nil || len(request) != 8 || binary.BigEndian.Uint16(request[:2]) != 1 ||
		binary.BigEndian.Uint32(request[2:6]) != 7 {
		t.Fatalf("EncodeRequest() = %v, %v", request, err)
	}
	response := make([]byte, 11)
	binary.BigEndian.PutUint16(response[:2], uint16(commonpb.MessageID_LoginPlayerRes))
	binary.BigEndian.PutUint32(response[2:6], 7)
	binary.BigEndian.PutUint32(response[6:10], uint32(commonpb.ErrorCode_ERROR_CODE_GATEWAY_TOKEN_INVALID))
	response[10] = 3
	decoded, err := DecodeInbound(response)
	if err != nil || decoded.Sequence != 7 || decoded.ErrorCode != commonpb.ErrorCode_ERROR_CODE_GATEWAY_TOKEN_INVALID ||
		len(decoded.Body) != 1 || decoded.Body[0] != 3 {
		t.Fatalf("DecodeInbound() = %+v, %v", decoded, err)
	}
}

func TestDecodeInboundPushUsesSixByteHeader(t *testing.T) {
	push := make([]byte, 7)
	binary.BigEndian.PutUint16(push[:2], uint16(commonpb.MessageID_PlayerKicked))
	push[6] = 9
	decoded, err := DecodeInbound(push)
	if err != nil || decoded.Sequence != 0 || len(decoded.Body) != 1 || decoded.Body[0] != 9 {
		t.Fatalf("DecodeInbound(push) = %+v, %v", decoded, err)
	}
}
