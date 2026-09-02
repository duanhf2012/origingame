// Package token 负责 Gateway 本地验证 LoginService 签发的游戏 Token。
package token

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"origingame/internal/security"
)

// Config 保存 Gateway 验签所需的公开材料。
type Config struct {
	Issuer     string            `json:"issuer"`
	Audience   string            `json:"audience"`
	PublicKeys map[string]string `json:"public_keys"`
}

// Verifier 使用 kid 选择 Ed25519 公钥，并严格校验标准 Claims。
type Verifier struct {
	issuer   string
	audience string
	keys     map[string]ed25519.PublicKey
}

// NewVerifier 校验并冻结公钥配置；错误信息不会包含密钥内容。
func NewVerifier(config Config) (*Verifier, error) {
	// 规范化配置中的标识字段，避免空白字符改变校验语义。
	config.Issuer = strings.TrimSpace(config.Issuer)
	config.Audience = strings.TrimSpace(config.Audience)
	if config.Issuer == "" || config.Audience == "" || len(config.PublicKeys) == 0 {
		return nil, errors.New("token issuer、audience 和 public_keys 不能为空")
	}
	// 解码并复制每把公钥，冻结验证器持有的配置。
	keys := make(map[string]ed25519.PublicKey, len(config.PublicKeys))
	for kid, encoded := range config.PublicKeys {
		kid = strings.TrimSpace(kid)
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
		if err != nil || kid == "" || len(raw) != ed25519.PublicKeySize {
			return nil, errors.New("token public_keys 包含无效 Ed25519 公钥")
		}
		keys[kid] = append(ed25519.PublicKey(nil), raw...)
	}
	return &Verifier{issuer: config.Issuer, audience: config.Audience, keys: keys}, nil
}

// Verify 返回签名可信且非空的 AccountID；失败时不区分内部原因给客户端。
func (verifier *Verifier) Verify(raw string) (string, error) {
	// 空 Token 或未初始化验证器直接按无效凭证处理。
	if verifier == nil || strings.TrimSpace(raw) == "" {
		return "", errors.New("token 无效")
	}
	// 使用固定算法、签发方、受众和有效期解析 Claims。
	claims := &security.GameClaims{}
	parsed, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(token *jwt.Token) (any, error) {
			// 拒绝算法降级，并按 kid 选择已登记公钥。
			if token.Method != jwt.SigningMethodEdDSA || token.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
				return nil, errors.New("token alg 无效")
			}
			kid, ok := token.Header["kid"].(string)
			if !ok || strings.TrimSpace(kid) == "" {
				return nil, errors.New("token kid 无效")
			}
			key, exists := verifier.keys[kid]
			if !exists {
				return nil, errors.New("token kid 未登记")
			}
			return key, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithIssuer(verifier.issuer),
		jwt.WithAudience(verifier.audience),
		jwt.WithExpirationRequired(),
	)
	// 仅返回已验签且携带账号主体的 Token。
	if err != nil || !parsed.Valid || strings.TrimSpace(claims.Subject) == "" {
		return "", errors.New("token 无效")
	}
	return strings.TrimSpace(claims.Subject), nil
}
