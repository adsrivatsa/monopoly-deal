-- name: GetGameTimeoutForUpdate :one
SELECT * FROM game_timeout WHERE game_id = $1 AND player_id = $2 AND demand_id IS NOT DISTINCT FROM $3 FOR UPDATE;

-- name: UpsertGameMoveTimeout :one
INSERT INTO game_timeout (game_id, player_id, token_id, duration_ms, deadline) VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (game_id, player_id) WHERE demand_id IS NULL DO UPDATE SET
    token_id = EXCLUDED.token_id,
    duration_ms = EXCLUDED.duration_ms,
    deadline = EXCLUDED.deadline
    RETURNING *;

-- name: UpsertGameDemandTimeout :one
INSERT INTO game_timeout (game_id, player_id, demand_id, token_id, duration_ms, deadline) VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (game_id, player_id, demand_id) WHERE demand_id IS NOT NULL DO UPDATE SET
    token_id = EXCLUDED.token_id,
    duration_ms = EXCLUDED.duration_ms,
    deadline = EXCLUDED.deadline
    RETURNING *;

-- name: DeleteGameTimeout :one
DELETE FROM game_timeout WHERE game_id = $1 AND player_id = $2 and demand_id IS NOT DISTINCT FROM $3 RETURNING *;

-- name: ListGameTimeouts :many
SELECT * FROM game_timeout WHERE game_id = $1 ORDER BY deadline;

-- name: ClaimGameTimeouts :many
UPDATE game_timeout SET claimed_at = NOW()
 WHERE (game_id, player_id, COALESCE(demand_id, '')) IN (
     SELECT game_id, player_id, COALESCE(demand_id, '') FROM game_timeout
      WHERE claimed_at IS NULL OR claimed_at < NOW() - INTERVAL '30 seconds'
      ORDER BY deadline
     LIMIT $1
    FOR UPDATE SKIP LOCKED
    )
    RETURNING *;