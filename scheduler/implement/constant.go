package implement

import (
	"github.com/chencheng8888/GoDo/dao/model"
	"github.com/chencheng8888/GoDo/scheduler/domain"
)

const registerTaskScriptStr = `
local jobKey = KEYS[1]
local zkey = KEYS[2]
local score = ARGV[1]
local member = ARGV[2]
local payload = ARGV[3]

local added = redis.call("ZADD", zkey, "NX", score, member)
if added == 0 then
  return 0
end

redis.call("HSET", jobKey, "payload", payload)
return 1
`

const removeTaskScriptStr = `
local jobKey = KEYS[1]
local zkey = KEYS[2]
local member = ARGV[1]

local removed = redis.call("ZREM", zkey, member)
if removed == 0 then
  return 0
end

redis.call("DEL", jobKey)
return 1
`

const getTaskScriptStr = `
-- pop_with_payload_by_score_le.lua
-- KEYS[1] = zkey
-- ARGV[1] = maxScore (inclusive, 如 now 的 unix ms 字符串)
-- ARGV[2] = limit (可选，返回最多多少个)
local zkey = KEYS[1]
local maxScore = ARGV[1]
local limit = tonumber(ARGV[2]) or 100

-- 获取 member + score 对（WITHSCORES 返回交替数组）
local res = redis.call('ZRANGEBYSCORE', zkey, '-inf', maxScore, 'LIMIT', 0, limit, 'WITHSCORES')
if #res == 0 then
  return {}
end

-- 抽出 members 用于批量 ZREM，同时组装返回结果（member, score, payload）
local members = {}
local out = {}
for i = 1, #res, 2 do
  local member = res[i]
  local score = res[i+1]
  table.insert(members, member)

  local payload = redis.call('HGET', member, 'payload')
  if not payload then payload = "" end

  -- append payload
  table.insert(out, payload)
end

-- 删除这些 member（批量删除）
if #members > 0 then
  redis.call('ZREM', zkey, unpack(members))
end

-- 返回 payload 列表
return out`

func newModel(task domain.Task) *model.TaskInfo {
	return &model.TaskInfo{
		TaskId:        task.GetID(),
		TaskName:      task.GetTaskName(),
		OwnerName:     task.GetOwnerName(),
		ScheduledTime: task.GetScheduledTime(),
		Description:   task.GetDescription(),
		JobType:       task.GetJob().Type(),
		Job:           task.GetJob().ToJson(),
	}
}
