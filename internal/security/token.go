// Package security 提供 OriginGame 各入口服务共用的安全基础能力。
package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenVersion = 1

// TokenConfig 描述 LoginService 签发游戏 JWT 所需的不可变配置。
type TokenConfig struct {
	Issuer     string
	Audience   string
	Expire     time.Duration
	ActiveKID  string
	PrivateKey string
}

// GameClaims 是 Gateway 验签后使用的 OriginGame JWT Claims。
type GameClaims struct {
	Version      int32 `json:"ver"`
	PlatformType int32 `json:"pt"`
	jwt.RegisteredClaims
}

// TokenIssuer 使用 Ed25519 私钥签发只面向 Gateway 的短期游戏 Token。
type TokenIssuer struct {
	issuer     string
	audience   string
	expire     time.Duration
	activeKID  string
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	now        func() time.Time
}

// NewTokenIssuer 校验并冻结配置。PrivateKey 接受 Base64 编码的 32 字节 Seed 或 64 字节私钥。
func NewTokenIssuer(config TokenConfig) (*TokenIssuer, error) {
	// 先校验不会泄露密钥内容的普通配置。
	config.Issuer = strings.TrimSpace(config.Issuer)
	config.Audience = strings.TrimSpace(config.Audience)
	config.ActiveKID = strings.TrimSpace(config.ActiveKID)
	if config.Issuer == "" || config.Audience == "" || config.ActiveKID == "" {
		return nil, errors.New("token issuer、audience 和 active_kid 不能为空")
	}
	if config.Expire <= 0 {
		return nil, errors.New("token expire 必须为正数")
	}

	// 密钥错误只报告格式，不把配置值拼入错误。
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(config.PrivateKey))
	if err != nil {
		return nil, errors.New("token private_key 不是有效 Base64")
	}
	var privateKey ed25519.PrivateKey
	switch len(raw) {
	case ed25519.SeedSize:
		privateKey = ed25519.NewKeyFromSeed(raw)
	case ed25519.PrivateKeySize:
		privateKey = append(ed25519.PrivateKey(nil), raw...)
	default:
		return nil, fmt.Errorf("token private_key 解码后必须是 %d 或 %d 字节", ed25519.SeedSize, ed25519.PrivateKeySize)
	}
	publicKey := append(ed25519.PublicKey(nil), privateKey.Public().(ed25519.PublicKey)...)
	return &TokenIssuer{
		issuer: config.Issuer, audience: config.Audience, expire: config.Expire,
		activeKID: config.ActiveKID, privateKey: privateKey, publicKey: publicKey,
		now: time.Now,
	}, nil
}

// Issue 为稳定 AccountID 签发一条 JWT 字符串；调用方不得传入平台凭证等敏感数据。
func (issuer *TokenIssuer) Issue(accountID string, platformType int32) (string, error) {
	// Token 的 subject 是 MongoDB ObjectID 十六进制字符串，空值不能签发。
	accountID = strings.TrimSpace(accountID)
	if issuer == nil || accountID == "" {
		return "", errors.New("token issuer 或 account_id 无效")
	}

	// 每次签发生成独立 JTI，避免相同账号同一秒登录得到完全相同的 Token。
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", fmt.Errorf("生成 token jti: %w", err)
	}
	now := issuer.now().UTC()
	claims := GameClaims{
		Version: tokenVersion, PlatformType: platformType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: issuer.issuer, Subject: accountID,
			Audience: jwt.ClaimStrings{issuer.audience},
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(issuer.expire)),
			ID:        base64.RawURLEncoding.EncodeToString(jtiBytes),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = issuer.activeKID
	return token.SignedString(issuer.privateKey)
}

// PublicKeyBase64 返回可配置给 Gateway 的 Ed25519 公钥，不返回或派生暴露私钥。
func (issuer *TokenIssuer) PublicKeyBase64() string {
	if issuer == nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(issuer.publicKey)
}
