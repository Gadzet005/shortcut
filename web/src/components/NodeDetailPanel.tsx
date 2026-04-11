import type { NodeTraceResponse } from "../api/types";

interface Props {
  node: NodeTraceResponse;
  onClose: () => void;
}

export function NodeDetailPanel({ node, onClose }: Props) {
  const hasError = !!node.error;

  return (
    <div
      style={{
        width: 300,
        padding: 16,
        borderLeft: "1px solid #e5e7eb",
        background: "#fff",
        overflowY: "auto",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 16,
        }}
      >
        <h3 style={{ margin: 0, fontSize: 16 }}>Node Details</h3>
        <button
          onClick={onClose}
          style={{
            background: "none",
            border: "none",
            fontSize: 18,
            cursor: "pointer",
            color: "#6b7280",
          }}
        >
          x
        </button>
      </div>

      <Field label="Node ID" value={node.node_id} mono />
      {node.node_type && <Field label="Type" value={node.node_type} />}
      <Field label="Duration" value={`${node.duration_ms}ms`} />
      <Field
        label="Started"
        value={new Date(node.started_at).toLocaleTimeString()}
      />
      <Field
        label="Finished"
        value={new Date(node.finished_at).toLocaleTimeString()}
      />
      {(node.status_code ?? 0) > 0 && (
        <Field label="Status Code" value={String(node.status_code)} />
      )}
      {(node.retry_count ?? 0) > 0 && (
        <Field label="Retries" value={String(node.retry_count)} />
      )}
      {hasError && <Field label="Error" value={node.error!} error />}

      {node.dependencies && node.dependencies.length > 0 && (
        <div style={{ marginTop: 12 }}>
          <div style={{ fontSize: 12, color: "#6b7280", marginBottom: 4 }}>
            Dependencies
          </div>
          {node.dependencies.map((dep) => (
            <div
              key={dep.node_id}
              style={{
                fontSize: 13,
                fontFamily: "monospace",
                padding: "2px 0",
              }}
            >
              {dep.node_id}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function Field({
  label,
  value,
  mono,
  error,
}: {
  label: string;
  value: string;
  mono?: boolean;
  error?: boolean;
}) {
  return (
    <div style={{ marginBottom: 10 }}>
      <div style={{ fontSize: 12, color: "#6b7280" }}>{label}</div>
      <div
        style={{
          fontSize: 14,
          fontFamily: mono ? "monospace" : "inherit",
          color: error ? "#dc2626" : "#111827",
          wordBreak: "break-all",
        }}
      >
        {value}
      </div>
    </div>
  );
}
