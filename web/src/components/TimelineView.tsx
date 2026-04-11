import type { TraceResponse, NodeTraceResponse } from "../api/types";

interface Props {
  trace: TraceResponse;
  onNodeClick: (nodeId: string) => void;
  selectedNodeId: string | null;
}

const statusColors: Record<string, string> = {
  ok: "#22c55e",
  error: "#ef4444",
  retried: "#eab308",
};

function getNodeStatus(nt: NodeTraceResponse): string {
  if (nt.error) return "error";
  if ((nt.retry_count ?? 0) > 0) return "retried";
  return "ok";
}

export function TimelineView({ trace, onNodeClick, selectedNodeId }: Props) {
  const traceStart = new Date(trace.started_at).getTime();
  const traceEnd = new Date(trace.finished_at).getTime();
  const totalDuration = traceEnd - traceStart || 1;

  const sorted = [...trace.node_traces].sort(
    (a, b) => new Date(a.started_at).getTime() - new Date(b.started_at).getTime()
  );

  return (
    <div style={{ padding: "16px 0" }}>
      {sorted.map((nt) => {
        const start = new Date(nt.started_at).getTime();
        const end = new Date(nt.finished_at).getTime();
        const leftPct = ((start - traceStart) / totalDuration) * 100;
        const widthPct = Math.max(((end - start) / totalDuration) * 100, 0.5);
        const isSelected = nt.node_id === selectedNodeId;

        return (
          <div
            key={nt.node_id}
            onClick={() => onNodeClick(nt.node_id)}
            style={{
              display: "flex",
              alignItems: "center",
              marginBottom: 6,
              cursor: "pointer",
              background: isSelected ? "#f0f0f0" : "transparent",
              borderRadius: 4,
              padding: "4px 8px",
            }}
          >
            <div
              style={{
                width: 140,
                flexShrink: 0,
                fontSize: 13,
                fontWeight: 500,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
              title={nt.node_id}
            >
              {nt.node_id}
            </div>
            <div
              style={{
                flex: 1,
                position: "relative",
                height: 24,
                background: "#e5e7eb",
                borderRadius: 4,
              }}
            >
              <div
                style={{
                  position: "absolute",
                  left: `${leftPct}%`,
                  width: `${widthPct}%`,
                  height: "100%",
                  background: statusColors[getNodeStatus(nt)] ?? "#6b7280",
                  borderRadius: 4,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: 11,
                  color: "#fff",
                  fontWeight: 500,
                  overflow: "hidden",
                }}
              >
                {nt.duration_ms}ms
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
