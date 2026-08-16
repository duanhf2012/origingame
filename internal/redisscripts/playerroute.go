package redisscripts

const (
	RegisterGameServiceID    = "register_game_service_v1"
	SetGameServiceDrainingID = "set_game_service_draining_v1"
	AssignOrGetPlayerID      = "assign_or_get_player_v1"
	BeginPlayerLoadID        = "begin_player_load_v1"
	CompletePlayerLoginID    = "complete_player_login_v1"
	ReleasePlayerLoadID      = "release_player_load_v1"
	MarkPlayerResidentID     = "mark_player_resident_v1"
	BeginPlayerReleaseID     = "begin_player_release_v1"
	RenewPlayerRoutesID      = "renew_player_routes_v1"
)

const RegisterGameServiceSource = `
local info = KEYS[1]
local lease = KEYS[2]
local candidates = KEYS[3]
local service_name = ARGV[1]
local node_id = ARGV[2]
local node_session_id = ARGV[3]
local max_players = tonumber(ARGV[4])
local lease_ttl = tonumber(ARGV[5])
local assigned = tonumber(redis.call('HGET', info, 'assigned_count') or '0')
redis.call('HSET', info,
  'service_name', service_name,
  'node_id', node_id,
  'node_session_id', node_session_id,
  'state', 'READY',
  'max_players', max_players)
redis.call('HSETNX', info, 'loading_count', 0)
redis.call('HSETNX', info, 'online_count', 0)
redis.call('HSETNX', info, 'resident_count', 0)
redis.call('HSETNX', info, 'assigned_count', assigned)
redis.call('PSETEX', lease, lease_ttl, '1')
redis.call('ZADD', candidates, assigned, info)
return 1
`

const SetGameServiceDrainingSource = `
local info = KEYS[1]
local candidates = KEYS[2]
if redis.call('HGET', info, 'node_id') ~= ARGV[1] or
   redis.call('HGET', info, 'node_session_id') ~= ARGV[2] then
  return 0
end
redis.call('HSET', info, 'state', 'DRAINING')
redis.call('ZREM', candidates, info)
return 1
`

const AssignOrGetPlayerSource = `
local route = KEYS[1]
local candidates = KEYS[2]
local connection_id = ARGV[1]
local account_id = ARGV[2]
local show_area_id = ARGV[3]
local real_area_id = ARGV[4]
local route_ttl = tonumber(ARGV[5])
local assigning_timeout = tonumber(ARGV[6])
local loading_timeout = tonumber(ARGV[7])
local scan_limit = tonumber(ARGV[8])
local excluded = {}
for i = 9, #ARGV do excluded[ARGV[i]] = true end
local now_parts = redis.call('TIME')
local now = tonumber(now_parts[1]) * 1000 + math.floor(tonumber(now_parts[2]) / 1000)

local function instance_valid(info)
  if not info or info == '' then return false end
  if redis.call('HGET', info, 'state') ~= 'READY' then return false end
  if redis.call('EXISTS', info .. ':lease') == 0 then return false end
  local assigned = tonumber(redis.call('HGET', info, 'assigned_count') or '-1')
  local maximum = tonumber(redis.call('HGET', info, 'max_players') or '0')
  return assigned >= 0 and assigned < maximum
end

local state = redis.call('HGET', route, 'state')
if state then
  local info = redis.call('HGET', route, 'gs_instance_key')
  local updated = tonumber(redis.call('HGET', route, 'updated_at_ms') or '0')
  local alive = info and redis.call('EXISTS', info .. ':lease') == 1
  if (state == 'ONLINE' or state == 'RESIDENT') and alive then
    return {1, redis.call('HGET', route, 'gs_service') or '',
      redis.call('HGET', route, 'gs_node_id') or '',
      redis.call('HGET', route, 'gs_node_session_id') or '', state}
  end
  if state == 'LEAVING' then return {3, '', '', '', state} end
  if state == 'ASSIGNING' and alive and now - updated < assigning_timeout then
    return {3, '', '', '', state}
  end
  if state == 'LOADING' and alive and now - updated < loading_timeout then
    return {3, '', '', '', state}
  end
  if info and (state == 'ASSIGNING' or state == 'LOADING') then
    local loading = tonumber(redis.call('HGET', info, 'loading_count') or '0')
    local assigned = tonumber(redis.call('HGET', info, 'assigned_count') or '0')
    if loading > 0 then redis.call('HINCRBY', info, 'loading_count', -1) end
    if assigned > 0 then assigned = redis.call('HINCRBY', info, 'assigned_count', -1) end
    if instance_valid(info) then redis.call('ZADD', candidates, assigned, info) end
  end
  redis.call('DEL', route)
end

local entries = redis.call('ZRANGE', candidates, 0, scan_limit - 1)
for _, info in ipairs(entries) do
  if not excluded[info] then
    if instance_valid(info) then
      local assigned = redis.call('HINCRBY', info, 'assigned_count', 1)
      redis.call('HINCRBY', info, 'loading_count', 1)
      redis.call('ZADD', candidates, assigned, info)
      local service_name = redis.call('HGET', info, 'service_name') or ''
      local node_id = redis.call('HGET', info, 'node_id') or ''
      local node_session_id = redis.call('HGET', info, 'node_session_id') or ''
      redis.call('HSET', route,
        'real_area_id', real_area_id,
        'show_area_id', show_area_id,
        'account_id', account_id,
        'state', 'ASSIGNING',
        'gs_service', service_name,
        'gs_node_id', node_id,
        'gs_node_session_id', node_session_id,
        'gs_instance_key', info,
        'assignment_connection_id', connection_id,
        'updated_at_ms', now)
      redis.call('PEXPIRE', route, route_ttl)
      return {2, service_name, node_id, node_session_id, 'ASSIGNING'}
    else
      redis.call('ZREM', candidates, info)
    end
  end
end
return {4, '', '', '', ''}
`

const BeginPlayerLoadSource = `
local route = KEYS[1]
local info = KEYS[2]
local lease = KEYS[3]
if redis.call('EXISTS', lease) == 0 or redis.call('HGET', route, 'state') ~= 'ASSIGNING' or
   redis.call('HGET', route, 'gs_instance_key') ~= info or
   redis.call('HGET', route, 'gs_node_id') ~= ARGV[1] or
   redis.call('HGET', route, 'gs_node_session_id') ~= ARGV[2] or
   redis.call('HGET', route, 'assignment_connection_id') ~= ARGV[3] then
  return 0
end
local now_parts = redis.call('TIME')
local now = tonumber(now_parts[1]) * 1000 + math.floor(tonumber(now_parts[2]) / 1000)
redis.call('HSET', route, 'state', 'LOADING', 'updated_at_ms', now)
redis.call('PEXPIRE', route, tonumber(ARGV[4]))
return 1
`

const CompletePlayerLoginSource = `
local route = KEYS[1]
local info = KEYS[2]
local candidates = KEYS[3]
if redis.call('HGET', route, 'gs_instance_key') ~= info or
   redis.call('HGET', route, 'gs_node_id') ~= ARGV[1] or
   redis.call('HGET', route, 'gs_node_session_id') ~= ARGV[2] then
  return 0
end
local state = redis.call('HGET', route, 'state')
if state == 'LOADING' then
  if redis.call('HGET', route, 'assignment_connection_id') ~= ARGV[3] then return 0 end
  if tonumber(redis.call('HGET', info, 'loading_count') or '0') > 0 then
    redis.call('HINCRBY', info, 'loading_count', -1)
  end
  redis.call('HINCRBY', info, 'online_count', 1)
elseif state == 'RESIDENT' then
  if tonumber(redis.call('HGET', info, 'resident_count') or '0') > 0 then
    redis.call('HINCRBY', info, 'resident_count', -1)
  end
  redis.call('HINCRBY', info, 'online_count', 1)
elseif state ~= 'ONLINE' then
  return 0
end
local now_parts = redis.call('TIME')
local now = tonumber(now_parts[1]) * 1000 + math.floor(tonumber(now_parts[2]) / 1000)
redis.call('HSET', route, 'state', 'ONLINE', 'assignment_connection_id', '', 'updated_at_ms', now)
redis.call('PEXPIRE', route, tonumber(ARGV[4]))
local assigned = tonumber(redis.call('HGET', info, 'assigned_count') or '0')
redis.call('ZADD', candidates, assigned, info)
return 1
`

const ReleasePlayerLoadSource = `
local route = KEYS[1]
local info = KEYS[2]
local candidates = KEYS[3]
local state = redis.call('HGET', route, 'state')
if (state ~= 'ASSIGNING' and state ~= 'LOADING') or
   redis.call('HGET', route, 'gs_instance_key') ~= info or
   redis.call('HGET', route, 'gs_node_id') ~= ARGV[1] or
   redis.call('HGET', route, 'gs_node_session_id') ~= ARGV[2] or
   redis.call('HGET', route, 'assignment_connection_id') ~= ARGV[3] then
  return 0
end
if tonumber(redis.call('HGET', info, 'loading_count') or '0') > 0 then
  redis.call('HINCRBY', info, 'loading_count', -1)
end
local assigned = tonumber(redis.call('HGET', info, 'assigned_count') or '0')
if assigned > 0 then assigned = redis.call('HINCRBY', info, 'assigned_count', -1) end
redis.call('DEL', route)
if redis.call('HGET', info, 'state') == 'READY' and redis.call('EXISTS', info .. ':lease') == 1 then
  redis.call('ZADD', candidates, assigned, info)
end
return 1
`

const MarkPlayerResidentSource = `
local route = KEYS[1]
local info = KEYS[2]
local candidates = KEYS[3]
if redis.call('HGET', route, 'state') ~= 'ONLINE' or
   redis.call('HGET', route, 'gs_instance_key') ~= info or
   redis.call('HGET', route, 'gs_node_id') ~= ARGV[1] or
   redis.call('HGET', route, 'gs_node_session_id') ~= ARGV[2] then
  return 0
end
if tonumber(redis.call('HGET', info, 'online_count') or '0') > 0 then
  redis.call('HINCRBY', info, 'online_count', -1)
end
redis.call('HINCRBY', info, 'resident_count', 1)
local now_parts = redis.call('TIME')
local now = tonumber(now_parts[1]) * 1000 + math.floor(tonumber(now_parts[2]) / 1000)
redis.call('HSET', route, 'state', 'RESIDENT', 'updated_at_ms', now)
redis.call('PEXPIRE', route, tonumber(ARGV[3]))
local assigned = tonumber(redis.call('HGET', info, 'assigned_count') or '0')
redis.call('ZADD', candidates, assigned, info)
return 1
`

const BeginPlayerReleaseSource = `
local route = KEYS[1]
local info = KEYS[2]
local candidates = KEYS[3]
if redis.call('HGET', route, 'state') ~= 'RESIDENT' or
   redis.call('HGET', route, 'gs_instance_key') ~= info or
   redis.call('HGET', route, 'gs_node_id') ~= ARGV[1] or
   redis.call('HGET', route, 'gs_node_session_id') ~= ARGV[2] then
  return 0
end
if tonumber(redis.call('HGET', info, 'resident_count') or '0') > 0 then
  redis.call('HINCRBY', info, 'resident_count', -1)
end
local assigned = tonumber(redis.call('HGET', info, 'assigned_count') or '0')
if assigned > 0 then assigned = redis.call('HINCRBY', info, 'assigned_count', -1) end
local now_parts = redis.call('TIME')
local now = tonumber(now_parts[1]) * 1000 + math.floor(tonumber(now_parts[2]) / 1000)
redis.call('HSET', route, 'state', 'LEAVING', 'updated_at_ms', now)
redis.call('PEXPIRE', route, tonumber(ARGV[3]))
if redis.call('HGET', info, 'state') == 'READY' and redis.call('EXISTS', info .. ':lease') == 1 then
  redis.call('ZADD', candidates, assigned, info)
end
return 1
`

const RenewPlayerRoutesSource = `
local renewed = 0
local now_parts = redis.call('TIME')
local now = tonumber(now_parts[1]) * 1000 + math.floor(tonumber(now_parts[2]) / 1000)
for _, route in ipairs(KEYS) do
  local state = redis.call('HGET', route, 'state')
  if (state == 'ONLINE' or state == 'RESIDENT') and
     redis.call('HGET', route, 'gs_node_id') == ARGV[1] and
     redis.call('HGET', route, 'gs_node_session_id') == ARGV[2] then
    redis.call('HSET', route, 'updated_at_ms', now)
    redis.call('PEXPIRE', route, tonumber(ARGV[3]))
    renewed = renewed + 1
  end
end
return renewed
`
