package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var updateStateLua = redis.NewScript(`
	local current_seq = redis.call('HGET', KEYS[1], 'last_seq')
	local current_at = redis.call('HGET', KEYS[1], 'last_at')

	local new_seq = tonumber(ARGV[3])
	local new_at = ARGV[4]

	-- If current state exists, reject out-of-order state regression
	if current_seq and tonumber(current_seq) then
		if new_seq <= tonumber(current_seq) and new_at <= current_at then
			return 0
		end
	end

	redis.call('HSET', KEYS[1],
		'state', ARGV[1],
		'reason', ARGV[2],
		'last_seq', ARGV[3],
		'last_at', ARGV[4],
		'since', ARGV[5]
	)
	return 1
`)

type CachedMachineState struct {
	State   string    `json:"state"`
	Reason  string    `json:"reason"`
	LastSeq int64     `json:"last_seq"`
	LastAt  time.Time `json:"last_at"`
	Since   time.Time `json:"since"`
}

type StateStore struct {
	client redis.Cmdable
}

func NewStateStore(client redis.Cmdable) *StateStore {
	return &StateStore{client: client}
}

func (s *StateStore) UpdateMachineState(
	ctx context.Context,
	machineID string,
	state string,
	reason string,
	seq int64,
	occurredAt time.Time,
) (bool, error) {
	key := fmt.Sprintf("fp:m:%s:state", machineID)
	nowIso := occurredAt.UTC().Format(time.RFC3339Nano)

	res, err := updateStateLua.Run(ctx, s.client, []string{key},
		state,
		reason,
		seq,
		nowIso,
		nowIso,
	).Int()

	if err != nil {
		return false, fmt.Errorf("failed to execute state update Lua: %w", err)
	}

	return res == 1, nil
}

func (s *StateStore) GetMachineState(ctx context.Context, machineID string) (*CachedMachineState, error) {
	key := fmt.Sprintf("fp:m:%s:state", machineID)
	data, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil // not found
	}

	st := &CachedMachineState{
		State:  data["state"],
		Reason: data["reason"],
	}
	return st, nil
}

func (s *StateStore) SetHeartbeat(ctx context.Context, machineID string, ttl time.Duration) error {
	key := fmt.Sprintf("fp:m:%s:hb", machineID)
	return s.client.Set(ctx, key, "1", ttl).Err()
}
