package token

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	"origingame/internal/security"
)

func TestVerifierAcceptsIssuedTokenAndRejectsWrongAudience(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 1)
	}
	issuer, err := security.NewTokenIssuer(security.TokenConfig{
		Issuer: "issuer", Audience: "gateway", Expire: time.Hour,
		ActiveKID: "key-1", PrivateKey: base64.StdEncoding.EncodeToString(seed),
	})
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewVerifier(Config{
		Issuer: "issuer", Audience: "gateway",
		PublicKeys: map[string]string{"key-1": issuer.PublicKeyBase64()},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := issuer.Issue("account-1")
	if err != nil {
		t.Fatal(err)
	}
	accountID, err := verifier.Verify(raw)
	if err != nil || accountID != "account-1" {
		t.Fatalf("Verify() = %q, %v", accountID, err)
	}

	wrong, _ := NewVerifier(Config{
		Issuer: "issuer", Audience: "other",
		PublicKeys: map[string]string{"key-1": issuer.PublicKeyBase64()},
	})
	if _, err = wrong.Verify(raw); err == nil {
		t.Fatal("错误 audience 的 Token 不应通过")
	}
}

func BenchmarkVerifierVerify(b *testing.B) {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 1)
	}
	issuer, err := security.NewTokenIssuer(security.TokenConfig{
		Issuer: "issuer", Audience: "gateway", Expire: time.Hour,
		ActiveKID: "key-1", PrivateKey: base64.StdEncoding.EncodeToString(seed),
	})
	if err != nil {
		b.Fatal(err)
	}
	verifier, err := NewVerifier(Config{
		Issuer: "issuer", Audience: "gateway",
		PublicKeys: map[string]string{"key-1": issuer.PublicKeyBase64()},
	})
	if err != nil {
		b.Fatal(err)
	}
	raw, err := issuer.Issue("account-1")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := verifier.Verify(raw); err != nil {
			b.Fatal(err)
		}
	}
}
