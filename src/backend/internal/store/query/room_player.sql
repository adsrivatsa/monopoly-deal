-- name: CreateRoomPlayer :one
   INSERT INTO room_player (room_id, player_id, is_host)
   VALUES ($1, $2, $3)
RETURNING *;

-- name: ListRoomPlayers :many
SELECT rp.*, r.display_name AS room_display_name, r.capacity AS room_capacity, r.occupied AS room_occupied,
       r.created_at AS room_created_at, p.display_name AS host_display_name, p.image_url AS host_image_url,
       COUNT(*) OVER () AS total_count
  FROM room_player rp
           INNER JOIN room r
           ON r.room_id = rp.room_id
           INNER JOIN player p
           ON p.player_id = rp.player_id
 WHERE (r.room_id::text = sqlc.narg('search')::text OR r.room_id::text ILIKE '%' || sqlc.narg('search') || '%' OR
        p.player_id::text = sqlc.narg('search')::text OR p.player_id::text ILIKE '%' || sqlc.narg('search') || '%' OR
        r.display_name = sqlc.narg('search') OR r.display_name ILIKE '%' || sqlc.narg('search') || '%' OR
        p.display_name = sqlc.narg('search') OR p.display_name ILIKE '%' || sqlc.narg('search') || '%' OR
        sqlc.narg('search') IS NULL)
   AND rp.is_host
 ORDER BY r.display_name
 LIMIT $1 OFFSET $2;

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
 ORDER BY joined_at
 LIMIT 1;