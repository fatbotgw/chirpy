-- name: CreateChirp :one
INSERT INTO chirps (id,
                    created_at,
                    updated_at,
                    body,
                    userid
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

