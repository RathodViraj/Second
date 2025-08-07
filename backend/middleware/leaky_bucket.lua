-- leaky_bucket.lua
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])        -- leak rate (per second)
local capacity = tonumber(ARGV[3])    -- bucket capacity

-- Fetch bucket
local data = redis.call("HMGET", key, "last", "level")
local last = tonumber(data[1]) or now
local level = tonumber(data[2]) or 0

-- Leak calculation
local leaked = (now - last) * rate
level = math.max(0, level - leaked)

-- Add new request
if (level + 1) > capacity then
  return 0  -- limit exceeded
else
  level = level + 1
  redis.call("HMSET", key, "last", now, "level", level)
  redis.call("EXPIRE", key, 60)
  return 1  -- allowed
end
