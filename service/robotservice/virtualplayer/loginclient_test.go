package virtualplayer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginClientUsesStableIdentityAndSelectsTCPGateway(t *testing.T) {
	var received loginRequest
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := json.NewDecoder(request.Body).Decode(&received); err != nil {
			t.Error(err)
		}
		_ = json.NewEncoder(writer).Encode(loginResponse{
			Token: "sensitive", Areas: []areaInfo{{ShowAreaID: 1, Gates: []gateInfo{
				{Protocol: "kcp", Address: "127.0.0.1:9002"},
				{Protocol: "tcp", Address: "127.0.0.1:9001"},
			}}},
		})
	}))
	defer server.Close()
	client, err := NewLoginClient(LoginClientConfig{
		URL: server.URL, ShowAreaID: 1, PlatformType: 1, PlatformIDPrefix: "robot-", Workers: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	result, err := client.Login(context.Background(), 42)
	if err != nil || result.GatewayAddress != "127.0.0.1:9001" || result.Token != "sensitive" {
		t.Fatalf("Login() = %+v, %v", result, err)
	}
	if received.PlatformID != "robot-42" || received.PlatformType != 1 {
		t.Fatalf("request = %+v", received)
	}
}

func TestLoginClientRejectsMissingTargetArea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(writer).Encode(loginResponse{Token: "token", Areas: []areaInfo{{ShowAreaID: 2}}})
	}))
	defer server.Close()
	client, err := NewLoginClient(LoginClientConfig{
		URL: server.URL, ShowAreaID: 1, PlatformIDPrefix: "robot-", Workers: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err = client.Login(context.Background(), 1); err == nil {
		t.Fatal("missing target area was accepted")
	}
}
