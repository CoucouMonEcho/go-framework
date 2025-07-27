redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', tonumber(ARGV[2]) - tonumber(ARGV[3]))
local cnt = redis.call('ZCARD', KEYS[1])
if cnt >= tonumber(ARGV[1]) then
    return 'true'
else
    redis.call('ZADD', KEYS[1], tonumber(ARGV[2]), ARGV[2])
    redis.call('PEXPIRE', KEYS[1], ARGV[3])
    return 'false'
end