CREATE TABLE IF NOT EXISTS failures (
    request_id        text        PRIMARY KEY,
    namespace_id      text        NOT NULL,
    graph_id          text        NOT NULL,
    method            text        NOT NULL,
    path              text        NOT NULL,
    started_at        timestamptz NOT NULL,
    finished_at       timestamptz NOT NULL,
    duration_ms       bigint      NOT NULL,
    status            text        NOT NULL,
    error             text        NOT NULL DEFAULT '',
    node_traces       jsonb       NOT NULL DEFAULT '[]'::jsonb,
    ready_to_retry_at timestamptz NOT NULL,
    num_retry         bigint      NOT NULL DEFAULT 0,
    request_body      bytea
);

CREATE INDEX IF NOT EXISTS failures_graph_idx
    ON failures (namespace_id, graph_id, finished_at DESC);

CREATE INDEX IF NOT EXISTS failures_ready_idx
    ON failures (ready_to_retry_at)
    WHERE status IN ('pending', 'processing', 'failed');
