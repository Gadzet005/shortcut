import { Handle, Position, type NodeProps } from "@xyflow/react";
import type { GraphNodeData } from "../lib/layout";

const statusColors: Record<string, string> = {
  ok: "#22c55e",
  error: "#ef4444",
  retried: "#eab308",
  cached: "#8b5cf6",
};

export function GraphNode({ data }: NodeProps) {
  const nodeData = data as unknown as GraphNodeData;
  const bg = statusColors[nodeData.status] ?? "#6b7280";

  return (
    <div
      style={{
        background: bg,
        color: "#fff",
        borderRadius: 8,
        padding: "8px 12px",
        minWidth: 150,
        textAlign: "center",
        fontSize: 13,
        boxShadow: "0 2px 6px rgba(0,0,0,0.2)",
      }}
    >
      <Handle type="target" position={Position.Top} />
      <div style={{ fontWeight: 600 }}>{nodeData.label}</div>
      <div style={{ fontSize: 11, opacity: 0.85 }}>
        {nodeData.nodeType} &middot; {nodeData.durationMs}ms
      </div>
      {nodeData.cached && (
        <div style={{ fontSize: 10, marginTop: 2 }}>cached</div>
      )}
      {nodeData.retryCount > 0 && (
        <div style={{ fontSize: 10, marginTop: 2 }}>
          retries: {nodeData.retryCount}
        </div>
      )}
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}
