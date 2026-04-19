-- name: CreateRoomPlayer :one
   INSERT INTO room_player (room_id, player_id, is_host)
   VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteRoomPlayer :exec
DELETE
  FROM room_player
 WHERE room_id = $1
   AND player_id = $2;

-- name: UpdateRoomPlayerHost :one
   UPDATE room_player
      SET is_host = $1
    WHERE room_id = $2
      AND player_id = $3
RETURNING *;

-- name: GetRoomPlayer :one
SELECT *
  FROM room_player
 WHERE player_id = $1;

-- name: GetOldestRoomPlayer :one
SELECT *
  FROM room_player
 WHERE room_id = $1
   AND player_id <> @leaving_player_id
 ORDER BY joined_at
 LIMIT 1;

-- name: ToggleRoomPlayerIsReady :one
   UPDATE room_player
      SET is_ready = NOT is_ready
    WHERE room_id = $1
      AND player_id = $2
RETURNING *;

-- name: DeleteRoomPlayersByRoom :exec
DELETE
  FROM room_player
 WHERE room_id = $1;