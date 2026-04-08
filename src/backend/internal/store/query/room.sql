-- name: CreateRoom :one
INSERT INTO room(display_name, capacity)
    VALUES ($1, $2)
RETURNING
    *;

-- name: ListRooms :many
SELECT
    *
FROM
    room
WHERE (room_id = sqlc.narg('search_term')
    OR room_id ILIKE '%' || sqlc.narg('search_term') || '%'
    OR display_name = sqlc.narg('search_term')
    OR display_name ILIKE '%' || sqlc.narg('search_term') || '%'
    OR sqlc.narg('search_term') IS NULL)
AND room_status <> 'completed';

