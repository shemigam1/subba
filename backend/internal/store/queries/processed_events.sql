-- name: InsertProcessedEvent :one
-- Try to claim processing rights for (request_id, consumer). Returns the newly
-- inserted row on success, or nothing on conflict (event already claimed).
-- ON CONFLICT DO NOTHING means the INSERT is a no-op when the pair already exists;
-- the caller must check whether a row was actually inserted (rowsAffected == 1).
INSERT INTO processed_events (request_id, consumer, status)
VALUES ($1, $2, 'processing')
ON CONFLICT (request_id, consumer) DO NOTHING
RETURNING request_id, consumer, status, result, created_at, processed_at;

-- name: GetProcessedEvent :one
-- Fetch the current record for (request_id, consumer). Used by the idempotency
-- wrapper to check whether a previous run already completed successfully.
SELECT request_id, consumer, status, result, created_at, processed_at
FROM processed_events
WHERE request_id = $1
  AND consumer   = $2;

-- name: MarkProcessedEventDone :exec
-- Stamp the record as done after the handler commits successfully.
UPDATE processed_events
SET status       = 'done',
    processed_at = now()
WHERE request_id = $1
  AND consumer   = $2;

-- name: MarkProcessedEventFailed :exec
-- Stamp the record as failed so operators can query and alert on stuck events.
UPDATE processed_events
SET status       = 'failed',
    processed_at = now()
WHERE request_id = $1
  AND consumer   = $2;
