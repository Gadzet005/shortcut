import type { TraceResponse } from "../api/types";

interface Props {
  traces: TraceResponse[];
  selectedIndex: number;
  onSelect: (index: number) => void;
}

export function TraceSelector({ traces, selectedIndex, onSelect }: Props) {
  return (
    <div
      style={{
        marginBottom: 16,
        padding: "12px 16px",
        background: "#ffffff",
        border: "1px solid #e5e7eb",
        borderRadius: 8,
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 12,
          flexWrap: "wrap",
        }}
      >
        <span
          style={{
            fontSize: 13,
            fontWeight: 600,
            color: "#374151",
          }}
        >
          Traces ({traces.length}):
        </span>

        <div
          style={{
            display: "flex",
            gap: 8,
            flexWrap: "wrap",
          }}
        >
          {traces.map((trace, idx) => (
            <button
              key={trace.request_id}
              onClick={() => onSelect(idx)}
              style={{
                padding: "6px 12px",
                fontSize: 12,
                fontFamily: "monospace",
                border: `1px solid ${selectedIndex === idx ? "#2563eb" : "#d1d5db"}`,
                borderRadius: 6,
                background: selectedIndex === idx ? "#eff6ff" : "#ffffff",
                color: selectedIndex === idx ? "#1e40af" : "#374151",
                cursor: "pointer",
                transition: "all 0.2s",
              }}
              onMouseEnter={(e) => {
                if (selectedIndex !== idx) {
                  e.currentTarget.style.background = "#f9fafb";
                }
              }}
              onMouseLeave={(e) => {
                if (selectedIndex !== idx) {
                  e.currentTarget.style.background = "#ffffff";
                }
              }}
            >
              #{idx + 1} {new Date(trace.started_at).toLocaleTimeString()}{" "}
              ({trace.duration_ms}ms)
            </button>
          ))}
        </div>

        {/* Кнопки навигации */}
        <div style={{ marginLeft: "auto", display: "flex", gap: 8 }}>
          <button
            onClick={() => onSelect(Math.max(0, selectedIndex - 1))}
            disabled={selectedIndex === 0}
            style={{
              padding: "4px 10px",
              fontSize: 12,
              border: "1px solid #d1d5db",
              borderRadius: 4,
              background: "#ffffff",
              cursor: selectedIndex === 0 ? "not-allowed" : "pointer",
              opacity: selectedIndex === 0 ? 0.5 : 1,
            }}
          >
            ← Next
          </button>
          <button
            onClick={() => onSelect(Math.min(traces.length - 1, selectedIndex + 1))}
            disabled={selectedIndex === traces.length - 1}
            style={{
              padding: "4px 10px",
              fontSize: 12,
              border: "1px solid #d1d5db",
              borderRadius: 4,
              background: "#ffffff",
              cursor: selectedIndex === traces.length - 1 ? "not-allowed" : "pointer",
              opacity: selectedIndex === traces.length - 1 ? 0.5 : 1,
            }}
          >
            Previous →
          </button>
        </div>
      </div>
    </div>
  );
}
