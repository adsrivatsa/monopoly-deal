-- name: GetGameTimeoutForUpdate :one
SELECT * FROM game_timeout WHERE game_id = $1 AND player_id = $2 AND demand_id IS NOT DISTINCT FROM $3 FOR UPDATE;

-- name: UpsertGameMoveTimeout :one
INSERT INTO game_timeout (game_id, player_id, token_id) VALUES ($1, $2, $3)
    ON CONFLICT (game_id, player_id) WHERE demand_id IS NULL DO UPDATE SET
    token_id = EXCLUDED.token_id
    RETURNING *;

-- name: DeleteGameTimeout :one
DELETE FROM game_timeout WHERE game_id = $1 AND player_id = $2 and demand_id IS NOT DISTINCT FROM $3 RETURNING *;