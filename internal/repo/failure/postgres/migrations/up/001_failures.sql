CREATE TABLE IF NOT EXISTS GraphFailures (
    request_id TEXT PRIMARY KEY,
    namespace_id TEXT,
    graph_id TEXT,
    method TEXT,
    path TEXT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    duration_ms BIGINT,
    status TEXT,
    error TEXT,
    last_retry_at TIMESTAMP,
    step int,
    retry_count INT,
    last_retry_successful BOOLEAN
);

CREATE TABLE IF NOT EXISTS NodeResults (
    request_id TEXT,
    node_id TEXT,
    node_type TEXT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    duration_ms BIGINT,
    status_code INT,
    retry_count INT,
    error TEXT,
    PRIMARY KEY (request_id, node_id)
);

CREATE TABLE IF NOT EXISTS NodeResultsDependency (
    request_id TEXT,
    node_id TEXT,
    dependency TEXT,
    PRIMARY KEY (request_id, node_id, dependency)
);

CREATE INDEX IF NOT EXISTS idx_node_results_req_id ON NodeResults(request_id);
CREATE INDEX IF NOT EXISTS idx_node_results_dependency_req_id ON NodeResultsDependency(request_id);
