package failurepostgres

import (
	"context"
	"errors"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ failure.Repo = (*Repo)(nil)

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

type Repo struct {
	pool *pgxpool.Pool
}

const insertSQL = `
INSERT INTO failures (
	request_id, namespace_id, graph_id, method, path,
	started_at, finished_at, duration_ms, status, error,
	node_traces, ready_to_retry_at, num_retry, request_body
) VALUES (
	$1, $2, $3, $4, $5,
	$6, $7, $8, $9, $10,
	$11, $12, $13, $14
)
ON CONFLICT (request_id) DO UPDATE SET
	namespace_id = EXCLUDED.namespace_id,
	graph_id = EXCLUDED.graph_id,
	method = EXCLUDED.method,
	path = EXCLUDED.path,
	started_at = EXCLUDED.started_at,
	finished_at = EXCLUDED.finished_at,
	duration_ms = EXCLUDED.duration_ms,
	status = EXCLUDED.status,
	error = EXCLUDED.error,
	node_traces = EXCLUDED.node_traces,
	ready_to_retry_at = EXCLUDED.ready_to_retry_at,
	num_retry = EXCLUDED.num_retry,
	request_body = EXCLUDED.request_body
`

func (r *Repo) Save(ctx context.Context, f failure.Failure) error {
	traces, err := encodeNodeTraces(f.NodeTraces)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, insertSQL,
		f.RequestID,
		f.NamespaceID,
		f.GraphID,
		f.Method,
		f.Path,
		f.StartedAt,
		f.FinishedAt,
		f.DurationMs,
		string(f.Status),
		f.Error,
		traces,
		f.ReadyToRetryAt,
		f.NumRetry,
		f.RequestBody,
	)
	return err
}

const selectColumns = `
	request_id, namespace_id, graph_id, method, path,
	started_at, finished_at, duration_ms, status, error,
	node_traces, ready_to_retry_at, num_retry, request_body
`

func (r *Repo) GetByRequestID(ctx context.Context, requestID string) (failure.Failure, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+selectColumns+" FROM failures WHERE request_id = $1", requestID)
	return scanRow(row)
}

func (r *Repo) Delete(ctx context.Context, requestID string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM failures WHERE request_id = $1", requestID)
	return err
}

func (r *Repo) ListByGraph(ctx context.Context, namespaceID, graphID string) ([]failure.Failure, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT "+selectColumns+" FROM failures WHERE namespace_id = $1 AND graph_id = $2 ORDER BY finished_at DESC",
		namespaceID, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

const claimSQL = `
WITH ready AS (
	SELECT request_id FROM failures
	WHERE ready_to_retry_at <= NOW() AND status IN ('pending', 'processing', 'failed')
	ORDER BY ready_to_retry_at
	LIMIT $1
	FOR UPDATE SKIP LOCKED
)
UPDATE failures f
SET ready_to_retry_at = NOW() + ($2::bigint || ' microseconds')::interval,
	status = 'processing'
FROM ready
WHERE f.request_id = ready.request_id
RETURNING f.request_id, f.namespace_id, f.graph_id, f.method, f.path,
	f.started_at, f.finished_at, f.duration_ms, f.status, f.error,
	f.node_traces, f.ready_to_retry_at, f.num_retry, f.request_body
`

func (r *Repo) ClaimReadyBatch(ctx context.Context, batchSize int, visibilityTimeout time.Duration) ([]failure.Failure, error) {
	rows, err := r.pool.Query(ctx, claimSQL, batchSize, visibilityTimeout.Microseconds())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repo) UpdateProgress(ctx context.Context, requestID string, numRetry int64, readyToRetryAt time.Time, status failure.Status) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE failures SET num_retry = $1, ready_to_retry_at = $2, status = $3 WHERE request_id = $4",
		numRetry, readyToRetryAt, string(status), requestID,
	)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRow(r rowScanner) (failure.Failure, error) {
	var (
		requestID, namespaceID, graphID, method, path string
		status, errorMsg                              string
		startedAt, finishedAt, readyToRetryAt         time.Time
		durationMs, numRetry                          int64
		nodeTraces                                    []byte
		requestBody                                   []byte
	)
	if err := r.Scan(
		&requestID, &namespaceID, &graphID, &method, &path,
		&startedAt, &finishedAt, &durationMs, &status, &errorMsg,
		&nodeTraces, &readyToRetryAt, &numRetry, &requestBody,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return failure.Failure{}, failure.ErrNotFound
		}
		return failure.Failure{}, err
	}
	return scanFailure(
		requestID, namespaceID, graphID, method, path, status, errorMsg,
		startedAt, finishedAt, readyToRetryAt,
		durationMs, numRetry,
		nodeTraces, requestBody,
	)
}

func scanRows(rows pgx.Rows) ([]failure.Failure, error) {
	var out []failure.Failure
	for rows.Next() {
		f, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
