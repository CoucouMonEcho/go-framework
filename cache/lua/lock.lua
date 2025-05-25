local val = redis.call('GET', KEYS[1])
if not val then
    return redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])
elseif val == ARGV[1] then
    redis.call('EXPIRE', KEYS[1], ARGV[2])
    return 'OK'
else
    return ''
end