import type { TraceResponse } from "../api/types";

interface Props {
  trace: TraceResponse;
}

export function TraceSummary({ trace }: Props) {
  const isError = trace.status === "error";

  return (
    <div
      style={{
        display: "flex",
        flexWrap: "wrap",
        gap: 16,
        alignItems: "center",
        padding: "12px 16px",
        background: "#f9fafb",
        borderRadius: 8,
        border: "1px solid #e5e7eb",
      }}
    >
      <div>
        <span style={{ fontSize: 12, color: "#6b7280" }}>Request ID</span>
        <div style={{ fontWeight: 600, fontSize: 14, fontFamily: "monospace" }}>
          {trace.request_id}
        </div>
      </div>
      <div>
        <span style={{ fontSize: 12, color: "#6b7280" }}>Graph</span>
        <div style={{ fontWeight: 500, fontSize: 14 }}>
          {trace.namespace_id}/{trace.graph_id}
        </div>
      </div>
      <div>
        <span style={{ fontSize: 12, color: "#6b7280" }}>Request</span>
        <div style={{ fontWeight: 500, fontSize: 14 }}>
          {trace.method} {trace.path}
        </div>
      </div>
      <div>
        <span style={{ fontSize: 12, color: "#6b7280" }}>Duration</span>
        <div style={{ fontWeight: 600, fontSize: 14 }}>
          {trace.duration_ms}ms
        </div>
      </div>
      <div>
        <span
          style={{
            display: "inline-block",
            padding: "2px 10px",
            borderRadius: 12,
            fontSize: 13,
            fontWeight: 600,
            background: isError ? "#fef2f2" : "#f0fdf4",
            color: isError ? "#dc2626" : "#16a34a",
            border: `1px solid ${isError ? "#fecaca" : "#bbf7d0"}`,
          }}
        >
          {trace.status}
        </span>
      </div>
      {trace.error && (
        <div style={{ width: "100%", color: "#dc2626", fontSize: 13 }}>
          {trace.error}
        </div>
      )}
    </div>
  );
}
