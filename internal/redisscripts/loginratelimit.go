// Package redisscripts 保存由 AccDBService 启动时登记的受控 Redis Script。
package redisscripts

const (
	// LoginRateLimitID 是 LoginService 滑动窗口限流使用的稳定登记 ID。
	LoginRateLimitID = "login_rate_limit_v1"

	// LoginRateLimitSource 原子清理过期记录、检查计数并登记当前请求。
	LoginRateLimitSource = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local maximum = tonumber(ARGV[3])
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
local count = redis.call('ZCOUNT', key, now - window, '+inf')
if count >= maximum then
  return 0
end
local member = tostring(now) .. '-' .. tostring(count)
redis.call('ZADD', key, now, member)
redis.call('PEXPIRE', key, window)
return 1
`
)
