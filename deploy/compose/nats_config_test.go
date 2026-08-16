package compose

import (
	"os"
	"regexp"
	"testing"
)

func TestNATSMaxPayloadCoversOriginRPCEnvelope(t *testing.T) {
	content, err := os.ReadFile("nats.conf")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`(?m)^max_payload:\s*33MB\s*$`).Match(content) {
		t.Fatal("nats.conf max_payload 必须为 Origin 32M 业务载荷预留完整包络空间")
	}
}
