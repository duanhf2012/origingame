package security

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestTokenIssuerIssue 验证 Header、业务 Claims、受众和 Ed25519 签名形成同一契约。
func TestTokenIssuerIssue(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 1)
	}
	issuer, err := NewTokenIssuer(TokenConfig{
		Issuer: "origingame-login", Audience: "origingame-gateway", Expire: time.Hour,
		ActiveKID: "test-key", PrivateKey: base64.StdEncoding.EncodeToString(seed),
	})
	if err != nil {
		t.Fatalf("NewTokenIssuer() error = %v", err)
	}
	fixedNow := time.Date(2026, 8, 12, 20, 0, 0, 0, time.UTC)
	issuer.now = func() time.Time { return fixedNow }

	raw, err := issuer.Issue("0123456789abcdef01234567")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims := &GameClaims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return issuer.publicKey, nil
	}, jwt.WithAudience("origingame-gateway"), jwt.WithIssuer("origingame-login"), jwt.WithTimeFunc(func() time.Time {
		return fixedNow.Add(time.Minute)
	}))
	if err != nil || !parsed.Valid {
		t.Fatalf("ParseWithClaims() valid=%v error=%v", parsed.Valid, err)
	}
	if parsed.Header["kid"] != "test-key" || claims.Subject != "0123456789abcdef01234567" ||
		claims.NotBefore != nil || claims.ID != "" {
		t.Fatalf("unexpected token: header=%v claims=%+v", parsed.Header, claims)
	}
}

// TestNewTokenIssuerRejectsBadKey 确保错误配置在开放 HTTP 监听前被拒绝。
func TestNewTokenIssuerRejectsBadKey(t *testing.T) {
	_, err := NewTokenIssuer(TokenConfig{
		Issuer: "issuer", Audience: "audience", Expire: time.Hour,
		ActiveKID: "kid", PrivateKey: base64.StdEncoding.EncodeToString([]byte("short")),
	})
	if err == nil {
		t.Fatal("NewTokenIssuer() accepted an invalid private key")
	}
}
