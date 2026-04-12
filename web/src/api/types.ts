export interface TraceResponse {
  request_id: string;
  namespace_id: string;
  graph_id: string;
  method: string;
  path: string;
  started_at: string;
  finished_at: string;
  duration_ms: number;
  status: "ok" | "error";
  error?: string;
  node_traces: NodeTraceResponse[];
}

export interface NodeTraceResponse {
  node_id: string;
  node_type?: string;
  dependencies?: NodeDependencyResponse[];
  started_at: string;
  finished_at: string;
  duration_ms: number;
  status_code?: number;
  retry_count?: number;
  cached?: boolean;
  error?: string;
}

export interface NodeDependencyResponse {
  node_id: string;
}
