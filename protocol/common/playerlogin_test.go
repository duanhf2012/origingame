package commonpb

import "testing"

func TestLoginProtocolNumbersStayStable(t *testing.T) {
	if MessageID_Ok != 0 ||
		MessageID_LoginPlayerReq != 1 ||
		MessageID_LoginPlayerRes != 2 ||
		MessageID_PlayerHeartbeatReq != 3 ||
		MessageID_PlayerKicked != 4 {
		t.Fatal("MessageID 编号发生变化")
	}
	if ErrorCode_ERROR_CODE_GATEWAY_LOGGED_IN_ELSEWHERE != 2007 {
		t.Fatal("顶号错误码编号发生变化")
	}
}

func TestLoginPlayerMessagesExposeOnlyConfirmedFields(t *testing.T) {
	request := &LoginPlayerRequest{Token: "token", ShowAreaId: 9}
	if request.GetToken() != "token" || request.GetShowAreaId() != 9 {
		t.Fatalf("登录请求字段异常: %+v", request)
	}

	result := &LoginPlayerResult{RoleInfo: &RoleInfo{
		AccountId:   "account-1",
		ShowAreaId:  9,
		Nickname:    "player",
		Level:       1,
		CreatedAtMs: 123,
	}}
	if result.GetRoleInfo().GetAccountId() != "account-1" {
		t.Fatalf("登录结果字段异常: %+v", result)
	}

	kicked := &PlayerKickedNotify{Reason: ErrorCode_ERROR_CODE_GATEWAY_LOGGED_IN_ELSEWHERE}
	if kicked.GetReason() != ErrorCode_ERROR_CODE_GATEWAY_LOGGED_IN_ELSEWHERE {
		t.Fatalf("顶号原因异常: %+v", kicked)
	}
}
