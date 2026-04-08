-- name: CreateRoomPlayer :one
INSERT INTO room_player(room_id, player_id, is_host)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetRoomPlayerByPlayer :one
SELECT
    *
FROM
    room_player
WHERE
    player_id = $1;

-- name: GetRoomPlayers :many
SELECT
    rp.room_id,
    rp.joined_at,
    rp.is_host,
    p.player_id,
    p.display_name,
    p.email,
    p.image_url
FROM
    room_player rp
    INNER JOIN player p ON p.player_id = rp.player_id
WHERE
    rp.room_id = $1;

-- name: DeleteRoomPlayer :exec
DELETE FROM room_player
WHERE room_id = $1
    AND player_id = $2;

-- name: SetRoomHost :one
WITH oldest AS (
    SELECT
        room_id,
        player_id
    FROM (
        SELECT
            room_id,
            player_id,
            ROW_NUMBER() OVER (PARTITION BY room_id ORDER BY joined_at ASC) AS rn
    FROM
        room_player) t
    WHERE
        rn = 1)
UPDATE
    room_player rp
SET
    is_host =(rp.player_id = o.player_id)
FROM
    oldest o
WHERE
    rp.room_id = o.room_id
RETURNING
    rp.*;

